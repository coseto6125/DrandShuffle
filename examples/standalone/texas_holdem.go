package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/drand/go-clients/client"
	"github.com/drand/go-clients/client/http"
)

// 德州撲克遊戲狀態
type TexasHoldemGame struct {
	// 玩家手牌，每個玩家2張牌
	PlayerHands map[int][]Card

	// 公共牌（翻牌、轉牌、河牌）
	CommunityCards []Card

	// 使用的輪次號碼，用於驗證
	Round uint64

	// 遊戲局號，用於確保不同局次有不同的洗牌結果
	GameSessionID string
}

// Card 表示一張撲克牌
type Card struct {
	Suit  string // 花色
	Value string // 點數
}

// 初始化新的德州撲克遊戲
func NewTexasHoldemGame(numPlayers int, round uint64, gameSessionID string) (*TexasHoldemGame, error) {
	if numPlayers < 2 || numPlayers > 10 {
		return nil, fmt.Errorf("玩家數量必須在2到10之間")
	}

	var shuffledDeck []Card
	var err error
	var newRound uint64

	// 如果指定了輪次號碼，使用該輪次的隨機信標
	if round > 0 {
		shuffledDeck, err = GetShuffledDeckByRound(round, gameSessionID)
		if err != nil {
			return nil, fmt.Errorf("無法獲取洗牌後的牌組: %v", err)
		}
		newRound = round
	} else {
		// 否則使用最新的隨機信標
		shuffledDeck, newRound, err = GetShuffledDeck(gameSessionID)
		if err != nil {
			return nil, fmt.Errorf("無法獲取洗牌後的牌組: %v", err)
		}
	}

	// 初始化遊戲
	game := &TexasHoldemGame{
		PlayerHands:    make(map[int][]Card),
		CommunityCards: make([]Card, 0, 5),
		Round:          newRound,
		GameSessionID:  gameSessionID,
	}

	// 確保有足夠的牌
	requiredCards := numPlayers*2 + 5 // 每個玩家2張牌 + 5張公共牌
	if len(shuffledDeck) < requiredCards {
		return nil, fmt.Errorf("牌組長度不足，需要 %d 張牌，但只有 %d 張", requiredCards, len(shuffledDeck))
	}

	// 發牌：每個玩家2張牌
	cardIndex := 0
	for player := 0; player < numPlayers; player++ {
		game.PlayerHands[player] = shuffledDeck[cardIndex : cardIndex+2]
		cardIndex += 2
	}

	// 發公共牌：5張
	game.CommunityCards = shuffledDeck[cardIndex : cardIndex+5]

	return game, nil
}

// 顯示遊戲狀態
func (g *TexasHoldemGame) DisplayGame() {
	fmt.Printf("德州撲克遊戲 (輪次: %d, 遊戲局號: %s)\n\n", g.Round, g.GameSessionID)

	// 顯示玩家手牌
	for player, cards := range g.PlayerHands {
		fmt.Printf("玩家 %d 的手牌: %s%s, %s%s\n",
			player+1,
			cards[0].Suit, cards[0].Value,
			cards[1].Suit, cards[1].Value)
	}

	// 顯示公共牌
	fmt.Println("\n公共牌:")

	// 翻牌 (前3張)
	fmt.Print("翻牌: ")
	for i := 0; i < 3; i++ {
		fmt.Printf("%s%s ", g.CommunityCards[i].Suit, g.CommunityCards[i].Value)
	}
	fmt.Println()

	// 轉牌 (第4張)
	fmt.Printf("轉牌: %s%s\n", g.CommunityCards[3].Suit, g.CommunityCards[3].Value)

	// 河牌 (第5張)
	fmt.Printf("河牌: %s%s\n", g.CommunityCards[4].Suit, g.CommunityCards[4].Value)
}

// 獲取遊戲使用的輪次號碼
func (g *TexasHoldemGame) GetRound() uint64 {
	return g.Round
}

// 獲取遊戲使用的遊戲局號
func (g *TexasHoldemGame) GetGameSessionID() string {
	return g.GameSessionID
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

// GetShuffledDeck 返回使用最新drand隨機信標洗牌後的牌組
// gameSessionID 參數用於確保不同遊戲局次有不同的洗牌結果
// 返回洗好的牌組和使用的輪次號碼
func GetShuffledDeck(gameSessionID string) ([]Card, uint64, error) {
	// 獲取最新的隨機性和輪次號碼
	randomness, round, err := getDrandRandomness(0)
	if err != nil {
		return nil, 0, fmt.Errorf("無法獲取最新隨機性: %v", err)
	}

	// 創建足夠的隨機性
	hasher := sha256.New()
	hasher.Write(randomness)
	// 加入遊戲局號以確保不同局次有不同的洗牌結果
	hasher.Write([]byte(gameSessionID))
	extendedRandomness := hasher.Sum(randomness)

	// 初始化並洗牌
	deck := initializeDeck()
	shuffledDeck := shuffleDeck(deck, extendedRandomness)

	return shuffledDeck, round, nil
}

// GetShuffledDeckByRound 返回使用指定輪次drand隨機信標洗牌後的牌組
// gameSessionID 參數用於確保不同遊戲局次有不同的洗牌結果
func GetShuffledDeckByRound(round uint64, gameSessionID string) ([]Card, error) {
	// 獲取指定輪次的隨機性
	randomness, err := getDrandRandomnessByRound(round)
	if err != nil {
		return nil, fmt.Errorf("無法獲取輪次 %d 的隨機性: %v", round, err)
	}

	// 創建足夠的隨機性
	hasher := sha256.New()
	hasher.Write(randomness)
	// 加入遊戲局號以確保不同局次有不同的洗牌結果
	hasher.Write([]byte(gameSessionID))
	extendedRandomness := hasher.Sum(randomness)

	// 初始化並洗牌
	deck := initializeDeck()
	shuffledDeck := shuffleDeck(deck, extendedRandomness)

	return shuffledDeck, nil
}

// 獲取drand最新隨機信標
func getDrandRandomness(round uint64) ([]byte, uint64, error) {
	// 設定drand客戶端
	urls := []string{"https://api.drand.sh", "https://drand.cloudflare.com"}

	// 使用 quicknet 鏈的哈希值
	chainHash, err := hex.DecodeString("52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971")
	if err != nil {
		return nil, 0, fmt.Errorf("無法解碼鏈哈希: %v", err)
	}

	// 創建上下文，設置5秒超時
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 創建drand客戶端
	drandClient, err := client.New(
		client.From(http.ForURLs(ctx, nil, urls, chainHash)...),
		client.WithChainHash(chainHash),
	)

	if err != nil {
		return nil, 0, fmt.Errorf("無法創建drand客戶端: %v", err)
	}
	defer drandClient.Close()

	// 創建新的上下文用於獲取隨機信標，設置3秒超時
	getCtx, getCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer getCancel()

	// 獲取隨機信標
	result, err := drandClient.Get(getCtx, round)
	if err != nil {
		if round == 0 {
			return nil, 0, fmt.Errorf("無法獲取最新隨機信標 (超時或網絡錯誤): %v", err)
		}
		return nil, 0, fmt.Errorf("無法獲取輪次 %d 的隨機信標 (超時或網絡錯誤): %v", round, err)
	}

	return result.GetRandomness(), result.GetRound(), nil
}

// 獲取指定輪次的drand隨機信標
func getDrandRandomnessByRound(round uint64) ([]byte, error) {
	randomness, _, err := getDrandRandomness(round)
	return randomness, err
}

func main() {
	// 檢查命令行參數
	var round uint64 = 0
	var err error
	var gameSessionID string = generateSecureGameSessionID() // 使用安全的遊戲局號生成函數

	// 處理命令行參數
	if len(os.Args) > 1 {
		round, err = strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil {
			log.Fatalf("無效的輪次號碼: %v", err)
		}
	}

	if len(os.Args) > 2 {
		gameSessionID = os.Args[2]
	}

	// 創建一個4人的德州撲克遊戲
	game, err := NewTexasHoldemGame(4, round, gameSessionID)
	if err != nil {
		// 處理網絡錯誤
		if strings.Contains(err.Error(), "超時或網絡錯誤") {
			fmt.Println("警告: 無法連接到 drand 網絡，請檢查您的網絡連接。")
			fmt.Println("錯誤詳情:", err)
			fmt.Println("您可以稍後再試，或使用本地隨機源作為備用。")
			os.Exit(1)
		}
		log.Fatalf("無法創建遊戲: %v", err)
	}

	// 顯示遊戲狀態
	game.DisplayGame()

	// 輸出驗證信息
	fmt.Printf("\n遊戲使用的輪次號碼: %d\n", game.GetRound())
	fmt.Printf("遊戲使用的遊戲局號: %s\n", game.GetGameSessionID())
	fmt.Println("任何人都可以使用相同的輪次號碼和遊戲局號重現完全相同的發牌結果。")
	fmt.Printf("驗證命令: go run texas_holdem.go %d %s\n", game.GetRound(), game.GetGameSessionID())
}

// 生成加密安全的遊戲局號
func generateSecureGameSessionID() string {
	bytes := make([]byte, 16) // 128位隨機數
	_, err := rand.Read(bytes)
	if err != nil {
		// 如果無法生成隨機數，退回到時間戳（不理想但總比沒有好）
		return "game_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return "game_" + hex.EncodeToString(bytes)
}
