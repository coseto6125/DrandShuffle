# 基於 drand 區塊鏈的公平發牌系統

這個專案展示了如何使用 drand 分散式隨機信標（Distributed Randomness Beacon）來實現一個公平、透明且可驗證的撲克牌發牌系統，特別適合用於需要持續運行的遊戲服務。

## 概述

drand 是一個分散式隨機信標網絡，由多個獨立節點組成，這些節點共同產生可公開驗證的隨機數。這些隨機數具有以下特性：

1. **公開可驗證性**：任何人都可以驗證產生的隨機數是否正確。
2. **不可預測性**：在公開之前，沒有人能夠預測隨機數的值。
3. **不可操縱性**：沒有單一實體可以操縱或影響隨機數的生成。

這些特性使 drand 成為需要公平隨機性的應用（如撲克牌發牌）的理想選擇。

## 目錄結構

```
go_drand/
├── drand_shuffle/      # 核心庫
│   ├── drand_manager.go # drand 客戶端管理器
│   ├── shuffle.go      # 洗牌和卡片處理邏輯
│   └── shuffle_mock.go # 測試用的模擬實現
├── examples/           # 示例應用
│   ├── integrated/     # 使用 drand_shuffle 庫的集成實現
│   │   └── texas_holdem.go
│   └── standalone/     # 獨立實現（不依賴 drand_shuffle 庫）
│       ├── texas_holdem.go
│       └── server.go   # 持續運行的服務（獨立實現）
└── tests/              # 測試
    ├── advanced_test.go # 進階測試
    ├── core_test.go     # 核心功能測試
    └── standalone/      # 獨立測試
        ├── card_test.go
        ├── deck_test.go
        └── shuffle_test.go
```

## 兩種實現方式

本項目提供了兩種不同的實現方式，以滿足不同的使用場景。以下分別詳細說明這兩種實現方式。

---

## 集成實現 (Integrated Implementation)

### 概述

集成實現使用 `drand_shuffle` 庫作為核心組件，提供了一個封裝完善的解決方案，適合需要持續運行服務的場景。

### 架構

1. **DrandManager**：核心組件，負責管理與 drand 網絡的連接，每 3 秒獲取一次最新的隨機信標，並提供緩存機制。
2. **洗牌模組**：使用 Fisher-Yates 洗牌算法，結合 drand 隨機信標和遊戲局號生成公平的洗牌結果。
3. **示例應用**：展示如何在德州撲克等遊戲中使用 DrandManager。

### 工作原理

1. 服務啟動時，初始化 `DrandManager` 並開始後台獲取隨機信標。
2. 每 3 秒，`DrandManager` 從 drand 網絡獲取最新的隨機信標並緩存。
3. 當遊戲需要發牌時，使用最新的隨機信標和遊戲局號生成洗牌結果。
4. `DrandManager` 提供緩存機制，減少對 drand 網絡的請求次數。

### 使用方法

#### 運行德州撲克示例

```bash
cd examples/integrated
go run texas_holdem.go
```

或者指定輪次號碼和遊戲局號：

```bash
cd examples/integrated
go run texas_holdem.go 16173144 game_12345
```

#### 作為庫使用

```go
import (
    "go_drand/drand_shuffle"
)

// 獲取 DrandManager 實例
drandManager, err := drand_shuffle.GetDrandManager()
if err != nil {
    log.Fatalf("無法初始化 DrandManager: %v", err)
}

// 啟動後台獲取（如果尚未啟動）
drandManager.StartBackgroundFetching()
defer drandManager.Close()

// 獲取最新的隨機性和輪次號碼
randomness, round, err := drandManager.GetLatestRandomness()
if err != nil {
    log.Fatalf("無法獲取最新隨機性: %v", err)
}

// 使用隨機性進行洗牌
gameSessionID := "your_game_session_id"
shuffledDeck, _, err := drand_shuffle.GetShuffledDeck(gameSessionID)
if err != nil {
    log.Fatalf("無法獲取洗牌後的牌組: %v", err)
}

// 使用洗好的牌進行遊戲...
```

### 優勢

- 提供了封裝完善的解決方案，包括緩存和錯誤處理
- 自動後台獲取隨機信標，減少對 drand 網絡的請求次數
- 提供了簡單易用的 API，適合快速開發

---

## 獨立實現 (Standalone Implementation)

### 概述

獨立實現完全不依賴於 `drand_shuffle` 庫，直接使用 drand 客戶端庫，專為高併發環境設計，提供了更靈活的實現方式。

### 架構

1. **直接客戶端管理**：直接使用 drand 客戶端庫，不依賴 DrandManager，適合高併發環境。
2. **輕量級洗牌模組**：同樣使用 Fisher-Yates 洗牌算法，但以更靈活的方式實現。
3. **服務器示例**：展示如何在持續運行的服務中使用 drand 客戶端。
4. **德州撲克示例**：展示如何在德州撲克遊戲中使用獨立實現。

### 工作原理

1. 服務啟動時，直接初始化 drand 客戶端，不使用 `DrandManager`。
2. 使用更輕量級的方式管理 drand 客戶端，適合高併發環境。
3. 每 10 秒獲取一次最新的隨機信標，並使用它進行洗牌模擬。
4. 提供更靈活的實現，允許開發者自定義隨機性獲取邏輯。

### 使用方法

#### 啟動服務

```bash
cd examples/standalone
go run server.go
```

這將啟動一個持續運行的服務，每隔一段時間獲取一次最新的 drand 隨機信標。

#### 運行德州撲克示例

```bash
cd examples/standalone
go run texas_holdem.go
```

或者指定輪次號碼和遊戲局號：

```bash
cd examples/standalone
go run texas_holdem.go 16173144 game_12345
```

#### 作為庫使用

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

#### 在 WebSocket 服務中使用

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

// 在 WebSocket 處理函數中
func handleGame(conn *websocket.Conn) {
    // 生成唯一的遊戲局號
    gameSessionID := generateSecureGameSessionID()
    
    // 安全地獲取最新的隨機信標
    mu.RLock()
    randomness := latestRandomness
    round := latestRound
    mu.RUnlock()
    
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

### 優勢

- 專為高併發環境設計，使用更輕量級的方式管理 drand 客戶端
- 提供了更靈活的實現，適合需要自定義隨機性獲取邏輯的開發者
- 直接使用 drand 客戶端庫，提供了更清晰的示例，便於理解底層實現

---

## 選擇哪種實現？

- 如果您需要一個封裝完善的解決方案，包括緩存和錯誤處理，選擇**集成實現**
- 如果您需要在高併發環境中使用，或者需要更靈活的隨機性獲取方式，選擇**獨立實現**
- 如果您正在開發一個需要自定義隨機性處理邏輯的服務，**獨立實現**可能更適合您
- 如果您想要更直接地了解 drand 客戶端的使用方式，**獨立實現**提供了更清晰的示例

## 測試與驗證

本系統提供了多種方式來測試和驗證洗牌結果的公平性和正確性。

### 測試結構

測試分為以下幾類：

1. **核心測試**：位於 `tests` 目錄，測試核心庫的基礎功能，如卡片轉換、牌組初始化和洗牌算法。
   ```
   tests/
   ├── core_test.go     # 測試基礎功能，如卡片轉換、牌組初始化和洗牌算法
   └── advanced_test.go # 測試進階功能，如錯誤處理和洗牌結果的可重現性
   ```

2. **獨立測試**：位於 `tests/standalone` 目錄，包含完全獨立的測試，不依賴於核心庫。
   ```
   tests/standalone/
   ├── card_test.go   # 獨立測試卡片轉換功能
   ├── deck_test.go   # 獨立測試牌組初始化功能
   └── shuffle_test.go  # 獨立測試洗牌功能
   ```

### 運行測試

#### 運行獨立測試

```bash
go test -v ./tests/standalone
```

這些測試不依賴於核心庫，可以在任何環境中運行。

#### 運行核心測試

```bash
go test -v ./tests
```

注意：由於依賴問題，核心測試可能無法正常運行。在這種情況下，可以運行獨立測試來驗證基本功能。

### 驗證洗牌結果

要驗證洗牌結果的公平性和可重現性，可以使用德州撲克示例程序並提供相同的輪次號碼和遊戲局號。您可以選擇使用集成實現或獨立實現：

#### 使用集成實現驗證

```bash
cd examples/integrated
go run texas_holdem.go 16173144 game_12345
```

#### 使用獨立實現驗證

```bash
cd examples/standalone
go run texas_holdem.go 16173144 game_12345
```

多次運行相同的命令，應該會得到完全相同的洗牌和發牌結果，這證明了系統的確定性和可驗證性。

### 安全性驗證

為了驗證系統的安全性，可以進行以下測試：

1. **不同遊戲局號的獨立性**：使用相同的輪次號碼但不同的遊戲局號，應該會得到不同的洗牌結果。
2. **不同輪次的獨立性**：使用不同的輪次號碼但相同的遊戲局號，應該會得到不同的洗牌結果。
3. **網絡錯誤處理**：系統應該能夠優雅地處理網絡錯誤，例如通過斷開網絡連接來測試超時機制。

## 技術細節

### drand 網絡

本系統使用 [drand](https://drand.love/) 網絡的 quicknet 鏈作為隨機性來源。quicknet 鏈每 3 秒產生一個新的隨機信標。

鏈哈希：`52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971`

### 洗牌算法

本系統使用 Fisher-Yates 洗牌算法，這是一種無偏的洗牌算法，確保每種可能的牌序都有相同的概率出現。

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
A4: 遊戲局號應該是不可預測的，最好使用加密安全的隨機數生成器生成。在 Go 中，可以使用 `crypto/rand` 包生成隨機字節，然後轉換為字符串。

### Q5: 如果兩個不同的遊戲使用了相同的輪次號碼和遊戲局號，會發生什麼？
A5: 它們會得到完全相同的洗牌結果。這就是為什麼遊戲局號必須是唯一的，特別是在同一個平台上運行的不同遊戲之間。建議將遊戲ID或時間戳作為遊戲局號的一部分，以確保唯一性。

## 貢獻

歡迎貢獻！請隨時提交 Pull Request 或開 Issue。

## 許可證

本專案採用 MIT 許可證。 