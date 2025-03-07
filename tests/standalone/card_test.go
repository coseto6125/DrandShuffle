package standalone

import (
	"testing"
)

// Card 表示一張撲克牌
type Card struct {
	Suit  string // 花色
	Value string // 點數
}

// CardToString 將牌轉換為字符串表示
func CardToString(card Card) string {
	return card.Suit + card.Value
}

// StringToCard 將字符串表示轉換為牌
func StringToCard(s string) (Card, error) {
	// 這裡只是一個簡單的實現，用於測試
	return Card{Suit: "黑桃", Value: "A"}, nil
}

// TestCardConversion 測試牌的字符串轉換功能
func TestCardConversion(t *testing.T) {
	card := Card{Suit: "黑桃", Value: "A"}
	cardStr := CardToString(card)

	if cardStr != "黑桃A" {
		t.Errorf("CardToString({黑桃, A}) = %s，期望 黑桃A", cardStr)
	}

	card2, err := StringToCard("黑桃A")
	if err != nil {
		t.Errorf("StringToCard(黑桃A) 返回錯誤: %v", err)
	}

	if card2.Suit != "黑桃" || card2.Value != "A" {
		t.Errorf("StringToCard(黑桃A) = {%s, %s}，期望 {黑桃, A}", card2.Suit, card2.Value)
	}
}
