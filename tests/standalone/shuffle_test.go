package standalone

import (
	"crypto/sha256"
	"testing"
)

// ShuffleDeck 使用Fisher-Yates算法洗牌
func ShuffleDeck(deck []Card, randomness []byte) []Card {
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
		// 使用隨機字節的一部分作為隨機索引
		j := int(randomness[i%len(randomness)]) % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// TestShuffleDeck 測試洗牌功能
func TestShuffleDeck(t *testing.T) {
	// 創建一個標準牌組
	deck := InitializeDeck()

	// 創建隨機性
	randomness := []byte("test_randomness_for_shuffle")

	// 洗牌
	shuffled := ShuffleDeck(deck, randomness)

	// 檢查洗牌後的牌組大小是否仍為52張
	if len(shuffled) != 52 {
		t.Errorf("洗牌後牌組大小錯誤，期望 52，得到 %d", len(shuffled))
	}

	// 檢查洗牌後的牌組是否包含所有原始牌
	originalCards := make(map[string]bool)
	shuffledCards := make(map[string]bool)

	for _, card := range deck {
		originalCards[CardToString(card)] = true
	}

	for _, card := range shuffled {
		shuffledCards[CardToString(card)] = true
	}

	for card := range originalCards {
		if !shuffledCards[card] {
			t.Errorf("洗牌後缺少牌 %s", card)
		}
	}

	// 檢查洗牌是否改變了順序（這不是一個嚴格的測試，因為理論上洗牌後可能與原始順序相同，但概率極低）
	different := false
	for i := 0; i < len(deck); i++ {
		if CardToString(deck[i]) != CardToString(shuffled[i]) {
			different = true
			break
		}
	}

	if !different {
		t.Errorf("洗牌後順序與原始順序相同，洗牌可能沒有正常工作")
	}

	// 測試相同隨機性的洗牌結果是否一致
	shuffled2 := ShuffleDeck(deck, randomness)
	for i := 0; i < len(shuffled); i++ {
		if CardToString(shuffled[i]) != CardToString(shuffled2[i]) {
			t.Errorf("相同隨機性的洗牌結果不一致，位置 %d: %s vs %s",
				i, CardToString(shuffled[i]), CardToString(shuffled2[i]))
		}
	}
}
