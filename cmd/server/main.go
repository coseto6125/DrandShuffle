package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drand/go-clients/client"
	"github.com/drand/go-clients/client/http"
)

func main() {
	log.Println("啟動 drand 隨機信標服務...")

	// 初始化 drand 客戶端
	urls := []string{"https://api.drand.sh", "https://drand.cloudflare.com"}
	chainHash, err := hex.DecodeString("52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971")
	if err != nil {
		log.Fatalf("無法解碼鏈哈希: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	drandClient, err := client.New(
		client.From(http.ForURLs(ctx, nil, urls, chainHash)...),
		client.WithChainHash(chainHash),
	)
	if err != nil {
		log.Fatalf("無法創建 drand 客戶端: %v", err)
	}
	defer drandClient.Close()

	// 設置信號處理，優雅地關閉服務
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 每 10 秒獲取一次最新的隨機信標
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				result, err := drandClient.Get(ctx, 0)
				cancel()

				if err != nil {
					log.Printf("警告: 無法獲取最新隨機信標: %v", err)
					continue
				}

				randomness := result.GetRandomness()
				round := result.GetRound()
				log.Printf("當前最新輪次: %d, 隨機性前 8 字節: %x", round, randomness[:8])

				// 模擬遊戲發牌
				gameSessionID := "demo_game_" + time.Now().Format("150405")

				// 創建足夠的隨機性
				hasher := sha256.New()
				hasher.Write(randomness)
				// 加入遊戲局號以確保不同局次有不同的洗牌結果
				hasher.Write([]byte(gameSessionID))
				extendedRandomness := hasher.Sum(randomness)

				// 初始化並洗牌
				deck := initializeDeck()
				shuffledDeck := shuffleDeck(deck, extendedRandomness)

				log.Printf("模擬發牌: 第一張牌 %s%s, 最後一張牌 %s%s",
					shuffledDeck[0].Suit, shuffledDeck[0].Value,
					shuffledDeck[len(shuffledDeck)-1].Suit, shuffledDeck[len(shuffledDeck)-1].Value)
			case <-sigChan:
				return
			}
		}
	}()

	// 等待終止信號
	<-sigChan
	log.Println("收到終止信號，正在關閉服務...")
	log.Println("服務已關閉")
}

// 初始化標準52張撲克牌
func initializeDeck() []Card {
	suits := []string{"黑桃", "紅心", "方塊", "梅花"}
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

	deck := make([]Card, 0, len(suits)*len(values))

	for _, suit := range suits {
		for _, value := range values {
			deck = append(deck, Card{Suit: suit, Value: value})
		}
	}

	return deck
}

// Card 表示一張撲克牌
type Card struct {
	Suit  string // 花色
	Value string // 點數
}

// 使用Fisher-Yates算法洗牌
func shuffleDeck(deck []Card, randomness []byte) []Card {
	shuffled := make([]Card, len(deck))
	copy(shuffled, deck)

	// 確保有足夠的隨機字節
	if len(randomness) < 8 {
		// 擴展隨機性
		hasher := sha256.New()
		hasher.Write(randomness)
		randomness = hasher.Sum(nil)
	}

	// Fisher-Yates 洗牌算法
	for i := len(shuffled) - 1; i > 0; i-- {
		// 安全地獲取隨機索引
		pos := i % max(1, len(randomness)-8)
		j := int(binary.BigEndian.Uint64(randomness[pos:pos+8]) % uint64(i+1))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// 輔助函數
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
