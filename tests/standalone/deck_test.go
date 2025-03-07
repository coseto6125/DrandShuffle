package standalone

import (
	"testing"
)

// InitializeDeck 初始化標準52張撲克牌
func InitializeDeck() []Card {
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

// TestInitializeDeck 測試牌組初始化功能
func TestInitializeDeck(t *testing.T) {
	deck := InitializeDeck()

	// 檢查牌組大小是否為52張
	if len(deck) != 52 {
		t.Errorf("牌組大小錯誤，期望 52，得到 %d", len(deck))
	}

	// 檢查花色和點數
	suitCount := make(map[string]int)
	valueCount := make(map[string]int)

	for _, card := range deck {
		suitCount[card.Suit]++
		valueCount[card.Value]++
	}

	// 檢查花色數量
	expectedSuits := map[string]int{
		"黑桃": 13,
		"紅心": 13,
		"方塊": 13,
		"梅花": 13,
	}

	for suit, count := range expectedSuits {
		if suitCount[suit] != count {
			t.Errorf("花色 %s 的數量錯誤，期望 %d，得到 %d", suit, count, suitCount[suit])
		}
	}

	// 檢查點數數量
	expectedValues := map[string]int{
		"A": 4, "2": 4, "3": 4, "4": 4, "5": 4, "6": 4, "7": 4,
		"8": 4, "9": 4, "10": 4, "J": 4, "Q": 4, "K": 4,
	}

	for value, count := range expectedValues {
		if valueCount[value] != count {
			t.Errorf("點數 %s 的數量錯誤，期望 %d，得到 %d", value, count, valueCount[value])
		}
	}
}
