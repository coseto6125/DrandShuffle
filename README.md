# 基於 drand 區塊鏈的公平發牌系統

這個專案展示了如何使用 drand 分散式隨機信標（Distributed Randomness Beacon）來實現一個公平、透明且可驗證的撲克牌發牌系統，特別適合用於需要持續運行的遊戲服務。

## 概述

drand 是一個分散式隨機信標網絡，由多個獨立節點組成，這些節點共同產生可公開驗證的隨機數。這些隨機數具有以下特性：

1. **公開可驗證性**：任何人都可以驗證產生的隨機數是否正確。
2. **不可預測性**：在公開之前，沒有人能夠預測隨機數的值。
3. **不可操縱性**：沒有單一實體可以操縱或影響隨機數的生成。

這些特性使 drand 成為需要公平隨機性的應用（如撲克牌發牌）的理想選擇。

## 系統架構

本系統採用了持續運行的服務架構，主要組件包括：

1. **DrandManager**：核心組件，負責管理與 drand 網絡的連接，每 3 秒獲取一次最新的隨機信標，並提供緩存機制。
2. **洗牌模組**：使用 Fisher-Yates 洗牌算法，結合 drand 隨機信標和遊戲局號生成公平的洗牌結果。
3. **服務器**：持續運行的服務，維護 DrandManager 的生命週期，並處理遊戲請求。
4. **示例應用**：展示如何在德州撲克等遊戲中使用本系統。

### 目錄結構

```
go_drand/
├── cmd/
│   ├── server/         # 持續運行的服務
│   │   └── main.go
│   └── examples/       # 示例應用（備用）
│       └── texas_holdem.go
├── drand_shuffle/      # 核心庫
│   ├── drand_manager.go
│   └── shuffle.go
└── examples/           # 主要示例應用
    └── texas_holdem.go
```

## 工作原理

本系統的工作原理如下：

1. 服務啟動時，初始化 DrandManager 並開始後台獲取隨機信標。
2. 每 3 秒，DrandManager 從 drand 網絡獲取最新的隨機信標並緩存。
3. 當遊戲需要發牌時，使用最新的隨機信標和遊戲局號生成洗牌結果。
4. 整個過程是確定性的，這意味著給定相同的隨機信標和遊戲局號，洗牌和發牌的結果將始終相同。

## 使用方法

### 啟動服務

```bash
cd cmd/server
go run main.go
```

這將啟動一個持續運行的服務，每 3 秒獲取一次最新的 drand 隨機信標。

### 運行德州撲克示例

```bash
cd examples
go run texas_holdem.go
```

或者指定輪次號碼和遊戲局號：

```bash
cd examples
go run texas_holdem.go 16173144 game_12345
```

### 作為庫使用

本專案提供了一個可以直接導入到其他 Go 專案中的庫，特別適合用於德州撲克等遊戲。

```go
import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "log"
    "time"

    "github.com/drand/go-clients/client"
    "github.com/drand/go-clients/client/http"
)

// 初始化 drand 客戶端
urls := []string{"https://api.drand.sh", "https://drand.cloudflare.com"}
chainHash, err := hex.DecodeString("52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971")
if err != nil {
    log.Fatalf("無法解碼鏈哈希: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

drandClient, err := client.New(
    client.From(http.ForURLs(ctx, nil, urls, chainHash)...),
    client.WithChainHash(chainHash),
)
if err != nil {
    log.Fatalf("無法創建 drand 客戶端: %v", err)
}
defer drandClient.Close()

// 獲取最新的隨機信標
getCtx, getCancel := context.WithTimeout(context.Background(), 3*time.Second)
defer getCancel()

result, err := drandClient.Get(getCtx, 0)
if err != nil {
    log.Fatalf("無法獲取最新隨機信標: %v", err)
}

randomness := result.GetRandomness()
round := result.GetRound()

// 設定遊戲局號，確保不同局次有不同的洗牌結果
gameSessionID := generateSecureGameSessionID()

// 創建足夠的隨機性
hasher := sha256.New()
hasher.Write(randomness)
// 加入遊戲局號以確保不同局次有不同的洗牌結果
hasher.Write([]byte(gameSessionID))
extendedRandomness := hasher.Sum(randomness)

// 初始化並洗牌
deck := initializeDeck()
shuffledDeck := shuffleDeck(deck, extendedRandomness)

// 然後可以按照遊戲規則發牌
```

## 在 WebSocket 服務中使用

對於使用 WebSocket 的遊戲服務，可以這樣集成本系統：

```go
// 在服務啟動時初始化 drand 客戶端
urls := []string{"https://api.drand.sh", "https://drand.cloudflare.com"}
chainHash, err := hex.DecodeString("52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971")
if err != nil {
    log.Fatalf("無法解碼鏈哈希: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

drandClient, err := client.New(
    client.From(http.ForURLs(ctx, nil, urls, chainHash)...),
    client.WithChainHash(chainHash),
)
if err != nil {
    log.Fatalf("無法創建 drand 客戶端: %v", err)
}
defer drandClient.Close()

// 啟動後台獲取
go func() {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
            result, err := drandClient.Get(ctx, 0)
            cancel()
            
            if err != nil {
                log.Printf("警告: 無法獲取最新隨機信標: %v", err)
                continue
            }
            
            // 存儲最新的隨機信標
            latestRandomness := result.GetRandomness()
            latestRound := result.GetRound()
            
            // 更新緩存
            // ...
        }
    }
}()

// 在 WebSocket 處理函數中
func handleGame(conn *websocket.Conn) {
    // 生成唯一的遊戲局號
    gameSessionID := generateSecureGameSessionID()
    
    // 獲取最新的隨機信標
    randomness, round := getLatestRandomness()
    
    // 創建足夠的隨機性
    hasher := sha256.New()
    hasher.Write(randomness)
    // 加入遊戲局號以確保不同局次有不同的洗牌結果
    hasher.Write([]byte(gameSessionID))
    extendedRandomness := hasher.Sum(randomness)
    
    // 初始化並洗牌
    deck := initializeDeck()
    shuffledDeck := shuffleDeck(deck, extendedRandomness)
    
    // 使用洗好的牌進行遊戲...
}
```

## 安全性和公平性保證

本系統提供以下安全性和公平性保證：

1. **不可預測性**：由於 drand 隨機信標的不可預測性，沒有人能夠提前知道牌的分配。

2. **不可操縱性**：drand 網絡由多個獨立節點組成，沒有單一實體可以操縱隨機數的生成。

3. **公開可驗證性**：任何人都可以使用輪次號碼和遊戲局號來驗證發牌結果，確保發牌過程的公平性。

4. **透明性**：整個系統的代碼是開源的，任何人都可以審查算法和實現。

5. **局次隔離**：通過使用遊戲局號參數，確保不同局次的洗牌結果是不同的，即使使用相同的輪次號碼。這防止了預先獲取隨機信標進行預測的可能性。

6. **並發安全**：使用讀寫互斥鎖確保多個 goroutine 可以安全地訪問共享數據，適合在高並發的遊戲服務器中使用。

7. **網絡超時處理**：系統實現了網絡超時機制，當無法在指定時間內連接到 drand 網絡時，會返回明確的錯誤信息，而不是無限期等待。

## 技術細節

### drand 網絡

本系統使用 [drand](https://drand.love/) 網絡的 quicknet 鏈作為隨機性來源。quicknet 鏈每 3 秒產生一個新的隨機信標。

鏈哈希：`52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971`

### 洗牌算法

本系統使用 Fisher-Yates 洗牌算法，這是一種無偏的洗牌算法，確保每種可能的牌序都有相同的概率出現。

### 緩存機制

系統實現了緩存機制，將獲取的隨機信標存儲在內存中，以減少重複請求。緩存大小限制為 100 個信標，以避免內存無限增長。

### 依賴項

- Go 1.16 或更高版本
- github.com/drand/go-clients

## 安裝

1. 克隆此倉庫：

```bash
git clone https://github.com/yourusername/go_drand.git
cd go_drand
```

2. 安裝依賴：

```bash
go get github.com/drand/go-clients
```

## 常見問題 (FAQ)

### Q1: 為什麼使用 drand 而不是普通的隨機數生成器？
A1: drand 提供了公開可驗證的隨機性，這意味著任何人都可以驗證隨機數的生成過程是公平的。普通的隨機數生成器無法提供這種透明度和可驗證性，因此不適合需要高度公平性的場景，如線上賭博或撲克遊戲。

### Q2: drand 隨機信標每 3 秒產生一次，如果我在它產生完了才去取，流程是怎樣的？
A2: 本系統設計了一個持續運行的服務，它會在後台每 3 秒自動獲取最新的隨機信標並緩存。當遊戲需要發牌時，它會使用已經緩存的最新隨機信標，而不需要等待。為了防止攻擊者預先獲取這個值並預測洗牌結果，我們引入了「遊戲局號」參數。即使攻擊者知道了隨機信標的值，如果不知道遊戲局號，也無法預測洗牌結果。

### Q3: 為什麼使用「遊戲局號」而不是「房間ID」？
A3: 「遊戲局號」更準確地反映了這個參數的用途 - 它是針對特定一局遊戲的唯一識別符，而不是持久存在的房間。每開始一局新遊戲就應該生成一個新的遊戲局號，這樣可以確保每局遊戲的洗牌結果都是獨立且不可預測的。

### Q4: 如何生成安全的遊戲局號？
A4: 遊戲局號應該是不可預測的，最好使用加密安全的隨機數生成器生成。在 Go 中，可以使用 `crypto/rand` 包生成隨機字節，然後轉換為字符串。例如：
```go
import (
    "crypto/rand"
    "encoding/hex"
)

func generateSecureGameSessionID() string {
    bytes := make([]byte, 16) // 128位隨機數
    _, err := rand.Read(bytes)
    if err != nil {
        // 如果無法生成隨機數，退回到時間戳（不理想但總比沒有好）
        return "game_" + strconv.FormatInt(time.Now().UnixNano(), 10)
    }
    return "game_" + hex.EncodeToString(bytes)
}
```

### Q5: 如果兩個不同的遊戲使用了相同的輪次號碼和遊戲局號，會發生什麼？
A5: 它們會得到完全相同的洗牌結果。這就是為什麼遊戲局號必須是唯一的，特別是在同一個平台上運行的不同遊戲之間。建議將遊戲ID或時間戳作為遊戲局號的一部分，以確保唯一性。

### Q6: 系統如何處理 drand 網絡不可用的情況？
A6: 系統實現了網絡超時機制，當無法在指定時間內（默認為5秒）連接到 drand 網絡時，會返回明確的錯誤信息，而不是無限期等待。這使得調用者可以決定如何處理這種情況，例如：
1. 顯示錯誤信息並要求用戶稍後再試
2. 切換到備用的 drand 節點
3. 使用本地的隨機源作為備用（雖然這會降低可驗證性）
4. 實現重試機制，嘗試多次連接

### Q7: 如何在高並發環境中使用這個系統？
A7: 本系統設計了並發安全的機制，使用讀寫互斥鎖確保多個 goroutine 可以安全地訪問共享數據。在高並發環境中，您只需要初始化一個 drand 客戶端實例，然後在不同的 goroutine 中使用它。例如：

```go
// 在服務啟動時初始化一個 drand 客戶端
urls := []string{"https://api.drand.sh", "https://drand.cloudflare.com"}
chainHash, err := hex.DecodeString("52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971")
if err != nil {
    log.Fatalf("無法解碼鏈哈希: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

drandClient, err := client.New(
    client.From(http.ForURLs(ctx, nil, urls, chainHash)...),
    client.WithChainHash(chainHash),
)
if err != nil {
    log.Fatalf("無法創建 drand 客戶端: %v", err)
}
defer drandClient.Close()

// 使用互斥鎖保護共享數據
var (
    mu sync.RWMutex
    latestRandomness []byte
    latestRound uint64
)

// 啟動後台獲取
go func() {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
            result, err := drandClient.Get(ctx, 0)
            cancel()
            
            if err != nil {
                log.Printf("警告: 無法獲取最新隨機信標: %v", err)
                continue
            }
            
            // 更新共享數據
            mu.Lock()
            latestRandomness = result.GetRandomness()
            latestRound = result.GetRound()
            mu.Unlock()
        }
    }
}()

// 在不同的 goroutine 中使用
go func() {
    // 遊戲 1
    gameSessionID1 := generateSecureGameSessionID()
    
    // 安全地獲取最新的隨機信標
    mu.RLock()
    randomness := latestRandomness
    round := latestRound
    mu.RUnlock()
    
    // 使用隨機信標進行洗牌
    // ...
}()

go func() {
    // 遊戲 2
    gameSessionID2 := generateSecureGameSessionID()
    
    // 安全地獲取最新的隨機信標
    mu.RLock()
    randomness := latestRandomness
    round := latestRound
    mu.RUnlock()
    
    // 使用隨機信標進行洗牌
    // ...
}()
```

### Q8: 如何處理網絡超時問題？
A8: 系統設計了明確的網絡超時機制，當連接 drand 網絡時會設置超時限制（默認為5秒）。當超時發生時，系統會返回明確的錯誤信息，而不是無限期等待。這使得應用程序可以快速響應網絡問題，並採取適當的措施。

處理網絡超時的最佳實踐包括：
1. **設置合理的超時時間**：太短可能導致不必要的錯誤，太長會使用戶等待過久
2. **實現重試機制**：使用指數退避策略進行多次嘗試
3. **提供清晰的錯誤信息**：讓用戶知道發生了什麼問題
4. **準備備用方案**：例如使用備用的 drand 節點或本地隨機源

### Q9: 如何確保系統的安全性？
A9: 本系統採取了多種措施確保安全性：
1. **使用加密安全的隨機數生成器**：生成遊戲局號時使用 `crypto/rand`
2. **輸入驗證**：對花色和點數進行驗證，確保輸入數據的有效性
3. **邊界檢查**：對牌組長度進行檢查，確保有足夠的牌可以發給所有玩家
4. **並發安全**：使用互斥鎖確保多個 goroutine 可以安全地訪問共享數據
5. **錯誤處理**：明確處理各種錯誤情況，避免程序崩潰
6. **網絡超時**：設置網絡超時，避免無限期等待
7. **緩存清理**：定期清理舊的緩存，避免內存泄漏

### Q10: 如何在生產環境中部署這個系統？
A10: 在生產環境中部署時，建議：
1. **使用容器化**：將服務打包為 Docker 容器，便於部署和擴展
2. **實現監控**：監控 drand 客戶端的狀態，包括成功率、響應時間等
3. **實現日誌記錄**：記錄關鍵操作和錯誤，便於排查問題
4. **實現健康檢查**：定期檢查 drand 客戶端是否正常工作
5. **實現備用方案**：準備備用的 drand 節點或本地隨機源
6. **實現自動重啟**：當服務崩潰時自動重啟
7. **實現負載均衡**：在多個實例之間分配負載

## 貢獻

歡迎貢獻！請隨時提交 Pull Request 或開 Issue。

## 許可證

本專案採用 MIT 許可證。 