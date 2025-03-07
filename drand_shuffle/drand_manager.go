package drand_shuffle

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/drand/go-clients/client"
	"github.com/drand/go-clients/client/http"
)

// DrandManager 管理 drand 隨機信標的獲取和緩存
type DrandManager struct {
	client       client.Client
	latestBeacon *client.Result
	beaconCache  map[uint64]*client.Result
	mutex        sync.RWMutex
	stopChan     chan struct{}
	isRunning    bool
}

var (
	// 單例實例
	instance *DrandManager
	once     sync.Once
)

// GetDrandManager 返回 DrandManager 的單例實例
func GetDrandManager() (*DrandManager, error) {
	var initErr error
	once.Do(func() {
		instance = &DrandManager{
			beaconCache: make(map[uint64]*client.Result),
			stopChan:    make(chan struct{}),
		}
		initErr = instance.initialize()
	})
	return instance, initErr
}

// initialize 初始化 drand 客戶端
func (dm *DrandManager) initialize() error {
	// 設定 drand 客戶端
	urls := []string{"https://api.drand.sh", "https://drand.cloudflare.com"}

	// 使用 quicknet 鏈的哈希值
	chainHash, err := hex.DecodeString("52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971")
	if err != nil {
		return fmt.Errorf("無法解碼鏈哈希: %v", err)
	}

	// 創建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 創建 drand 客戶端
	dm.client, err = client.New(
		client.From(http.ForURLs(ctx, nil, urls, chainHash)...),
		client.WithChainHash(chainHash),
	)

	if err != nil {
		return fmt.Errorf("無法創建 drand 客戶端: %v", err)
	}

	// 獲取初始隨機信標
	err = dm.fetchLatestBeacon()
	if err != nil {
		return fmt.Errorf("無法獲取初始隨機信標: %v", err)
	}

	return nil
}

// StartBackgroundFetching 開始後台獲取隨機信標
func (dm *DrandManager) StartBackgroundFetching() {
	dm.mutex.Lock()
	if dm.isRunning {
		dm.mutex.Unlock()
		return
	}
	dm.isRunning = true
	dm.mutex.Unlock()

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := dm.fetchLatestBeacon()
				if err != nil {
					log.Printf("警告: 無法獲取最新隨機信標: %v", err)
				} else {
					log.Printf("成功獲取輪次 %d 的隨機信標", dm.latestBeacon.GetRound())
				}
			case <-dm.stopChan:
				return
			}
		}
	}()

	log.Println("已啟動後台 drand 隨機信標獲取服務")
}

// StopBackgroundFetching 停止後台獲取隨機信標
func (dm *DrandManager) StopBackgroundFetching() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if !dm.isRunning {
		return
	}

	dm.stopChan <- struct{}{}
	dm.isRunning = false
	log.Println("已停止後台 drand 隨機信標獲取服務")
}

// fetchLatestBeacon 獲取最新的隨機信標
func (dm *DrandManager) fetchLatestBeacon() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dm.client.Get(ctx, 0)
	if err != nil {
		return fmt.Errorf("無法獲取最新隨機信標: %v", err)
	}

	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// 檢查是否已經有這個輪次的信標
	if dm.latestBeacon != nil && dm.latestBeacon.GetRound() >= result.GetRound() {
		return nil // 已經有更新或相同的信標，不需要更新
	}

	dm.latestBeacon = result
	dm.beaconCache[result.GetRound()] = result

	// 清理舊的緩存，只保留最近 100 個
	const maxCacheSize = 100
	if len(dm.beaconCache) > maxCacheSize {
		// 收集所有輪次號碼
		rounds := make([]uint64, 0, len(dm.beaconCache))
		for round := range dm.beaconCache {
			rounds = append(rounds, round)
		}

		// 按輪次號碼排序
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})

		// 計算需要刪除的數量
		deleteCount := len(dm.beaconCache) - maxCacheSize

		// 刪除最舊的 deleteCount 個項目
		for i := 0; i < deleteCount; i++ {
			delete(dm.beaconCache, rounds[i])
		}

		log.Printf("已清理 %d 個舊的隨機信標緩存", deleteCount)
	}

	return nil
}

// GetLatestRandomness 獲取最新的隨機性和輪次號碼
func (dm *DrandManager) GetLatestRandomness() ([]byte, uint64, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.latestBeacon == nil {
		return nil, 0, fmt.Errorf("尚未獲取任何隨機信標")
	}

	return dm.latestBeacon.GetRandomness(), dm.latestBeacon.GetRound(), nil
}

// GetRandomnessByRound 獲取指定輪次的隨機性
func (dm *DrandManager) GetRandomnessByRound(round uint64) ([]byte, error) {
	dm.mutex.RLock()

	// 檢查緩存
	if beacon, ok := dm.beaconCache[round]; ok {
		randomness := beacon.GetRandomness()
		dm.mutex.RUnlock()
		return randomness, nil
	}
	dm.mutex.RUnlock()

	// 緩存中沒有，從網絡獲取
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dm.client.Get(ctx, round)
	if err != nil {
		return nil, fmt.Errorf("無法獲取輪次 %d 的隨機信標: %v", round, err)
	}

	// 更新緩存
	dm.mutex.Lock()
	dm.beaconCache[round] = result
	dm.mutex.Unlock()

	return result.GetRandomness(), nil
}

// Close 關閉 DrandManager
func (dm *DrandManager) Close() {
	if dm.isRunning {
		dm.StopBackgroundFetching()
	}
	if dm.client != nil {
		dm.client.Close()
	}
}
