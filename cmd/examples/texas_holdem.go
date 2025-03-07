package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go_drand/drand_shuffle"
)

// 德州撲克遊戲狀態
type TexasHoldemGame struct {
	// 玩家手牌，每個玩家2張牌
	PlayerHands map[int][]drand_shuffle.Card

	// 公共牌（翻牌、轉牌、河牌）
	CommunityCards []drand_shuffle.Card

	// 使用的輪次號碼，用於驗證
	Round uint64

	// 遊戲局號，用於確保不同局次有不同的洗牌結果
	GameSessionID string
}

// 初始化新的德州撲克遊戲
func NewTexasHoldemGame(numPlayers int, round uint64, gameSessionID string) (*TexasHoldemGame, error) {
	if numPlayers < 2 || numPlayers > 10 {
		return nil, fmt.Errorf("玩家數量必須在2到10之間")
	}

	var shuffledDeck []drand_shuffle.Card
	var err error
	var newRound uint64

	// 如果指定了輪次號碼，使用該輪次的隨機信標
	if round > 0 {
		shuffledDeck, err = drand_shuffle.GetShuffledDeckByRound(round, gameSessionID)
		if err != nil {
			return nil, fmt.Errorf("無法獲取洗牌後的牌組: %v", err)
		}
		newRound = round
	} else {
		// 否則使用最新的隨機信標
		shuffledDeck, newRound, err = drand_shuffle.GetShuffledDeck(gameSessionID)
		if err != nil {
			return nil, fmt.Errorf("無法獲取洗牌後的牌組: %v", err)
		}
	}

	// 初始化遊戲
	game := &TexasHoldemGame{
		PlayerHands:    make(map[int][]drand_shuffle.Card),
		CommunityCards: make([]drand_shuffle.Card, 0, 5),
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

func main() {
	// 初始化 DrandManager
	drandManager, err := drand_shuffle.GetDrandManager()
	if err != nil {
		log.Fatalf("無法初始化 DrandManager: %v", err)
	}

	// 啟動後台獲取（如果尚未啟動）
	drandManager.StartBackgroundFetching()
	defer drandManager.Close()

	// 檢查命令行參數
	var round uint64 = 0
	var gameSessionID string = generateSecureGameSessionID()

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
