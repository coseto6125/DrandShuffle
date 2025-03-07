package drand_shuffle

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

// Card 表示一張撲克牌
type Card struct {
	Suit  string // 花色
	Value string // 點數
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
	// 獲取 DrandManager 實例
	drandManager, err := GetDrandManager()
	if err != nil {
		return nil, 0, fmt.Errorf("無法初始化 DrandManager: %v", err)
	}

	// 獲取最新的隨機性和輪次號碼
	randomness, round, err := drandManager.GetLatestRandomness()
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
	// 獲取 DrandManager 實例
	drandManager, err := GetDrandManager()
	if err != nil {
		return nil, fmt.Errorf("無法初始化 DrandManager: %v", err)
	}

	// 獲取指定輪次的隨機性
	randomness, err := drandManager.GetRandomnessByRound(round)
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

// CardToString 將牌轉換為字符串表示
func CardToString(card Card) string {
	return card.Suit + card.Value
}

// StringToCard 將字符串表示轉換為牌
func StringToCard(s string) (Card, error) {
	if len(s) < 3 { // 至少需要3個字符：2個字符的花色 + 1個字符的點數
		return Card{}, fmt.Errorf("無效的牌字符串: %s", s)
	}

	// 驗證花色
	validSuits := []string{"黑桃", "紅心", "方塊", "梅花"}
	suit := ""
	value := ""

	// 嘗試匹配花色
	for _, validSuit := range validSuits {
		if strings.HasPrefix(s, validSuit) {
			suit = validSuit
			value = s[len(validSuit):]
			break
		}
	}

	if suit == "" {
		return Card{}, fmt.Errorf("無效的花色: %s", s)
	}

	// 驗證點數
	validValues := map[string]bool{
		"A": true, "2": true, "3": true, "4": true, "5": true,
		"6": true, "7": true, "8": true, "9": true, "10": true,
		"J": true, "Q": true, "K": true,
	}
	if !validValues[value] {
		return Card{}, fmt.Errorf("無效的點數: %s", value)
	}

	return Card{Suit: suit, Value: value}, nil
}

// LogDeck 打印牌組（用於調試）
func LogDeck(deck []Card) {
	for i, card := range deck {
		log.Printf("%d: %s%s\n", i+1, card.Suit, card.Value)
	}
}
