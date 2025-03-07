package tests

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"go_drand/drandshuffle"
)

// TestCardConversion 測試牌的字符串轉換功能
func TestCardConversion(t *testing.T) {
	testCases := []struct {
		name        string
		cardString  string
		expectedOk  bool
		expectedErr string
	}{
		{
			name:       "Valid card - 黑桃A",
			cardString: "黑桃A",
			expectedOk: true,
		},
		{
			name:       "Valid card - 紅心10",
			cardString: "紅心10",
			expectedOk: true,
		},
		{
			name:       "Valid card - 方塊J",
			cardString: "方塊J",
			expectedOk: true,
		},
		{
			name:       "Valid card - 梅花K",
			cardString: "梅花K",
			expectedOk: true,
		},
		{
			name:        "Invalid card - too short",
			cardString:  "黑A",
			expectedOk:  false,
			expectedErr: "無效的花色",
		},
		{
			name:        "Invalid card - invalid suit",
			cardString:  "星星A",
			expectedOk:  false,
			expectedErr: "無效的花色",
		},
		{
			name:        "Invalid card - invalid value",
			cardString:  "黑桃Z",
			expectedOk:  false,
			expectedErr: "無效的點數",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			card, err := drandshuffle.StringToCard(tc.cardString)

			if tc.expectedOk {
				assert.NoError(t, err)
				// 測試轉換回字符串
				cardString := drandshuffle.CardToString(card)
				assert.Equal(t, tc.cardString, cardString)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

// TestInitializeDeck 測試牌組初始化功能
func TestInitializeDeck(t *testing.T) {
	t.Run("Test deck initialization", func(t *testing.T) {
		// 創建一個標準牌組
		deck := drandshuffle.InitializeDeck()

		// 驗證牌組包含52張牌
		assert.Equal(t, 52, len(deck), "Deck should contain 52 cards")

		// 檢查所有牌是否都存在
		suitCounts := make(map[string]int)
		valueCounts := make(map[string]int)

		for _, card := range deck {
			suitCounts[card.Suit]++
			valueCounts[card.Value]++
		}

		assert.Equal(t, 13, suitCounts["黑桃"], "Should have 13 Spades")
		assert.Equal(t, 13, suitCounts["紅心"], "Should have 13 Hearts")
		assert.Equal(t, 13, suitCounts["方塊"], "Should have 13 Diamonds")
		assert.Equal(t, 13, suitCounts["梅花"], "Should have 13 Clubs")

		assert.Equal(t, 4, valueCounts["A"], "Should have 4 Aces")
		assert.Equal(t, 4, valueCounts["K"], "Should have 4 Kings")
		assert.Equal(t, 4, valueCounts["Q"], "Should have 4 Queens")
		assert.Equal(t, 4, valueCounts["J"], "Should have 4 Jacks")
		assert.Equal(t, 4, valueCounts["10"], "Should have 4 Tens")
		assert.Equal(t, 4, valueCounts["9"], "Should have 4 Nines")
		assert.Equal(t, 4, valueCounts["8"], "Should have 4 Eights")
		assert.Equal(t, 4, valueCounts["7"], "Should have 4 Sevens")
		assert.Equal(t, 4, valueCounts["6"], "Should have 4 Sixes")
		assert.Equal(t, 4, valueCounts["5"], "Should have 4 Fives")
		assert.Equal(t, 4, valueCounts["4"], "Should have 4 Fours")
		assert.Equal(t, 4, valueCounts["3"], "Should have 4 Threes")
		assert.Equal(t, 4, valueCounts["2"], "Should have 4 Twos")
	})
}

// TestShuffleDeck 測試洗牌功能
func TestShuffleDeck(t *testing.T) {
	t.Run("Test shuffle with different randomness", func(t *testing.T) {
		// 創建兩個不同的隨機數據
		randomBytes1 := make([]byte, 32)
		randomBytes2 := make([]byte, 32)

		_, err1 := rand.Read(randomBytes1)
		_, err2 := rand.Read(randomBytes2)

		assert.NoError(t, err1, "Failed to generate random bytes 1")
		assert.NoError(t, err2, "Failed to generate random bytes 2")

		// 創建一個標準牌組
		deck := drandshuffle.InitializeDeck()

		// 使用不同的隨機數據洗牌
		shuffled1 := drandshuffle.ShuffleDeck(deck, randomBytes1)
		shuffled2 := drandshuffle.ShuffleDeck(deck, randomBytes2)

		// 檢查兩次洗牌結果是否不同
		// 注意：理論上有極小的可能性兩次洗牌結果相同，但概率非常低
		different := false
		for i := 0; i < len(shuffled1); i++ {
			if drandshuffle.CardToString(shuffled1[i]) != drandshuffle.CardToString(shuffled2[i]) {
				different = true
				break
			}
		}

		assert.True(t, different, "Two shuffles with different randomness should produce different results")

		// 檢查洗牌後的牌組仍然包含所有52張牌
		assert.Equal(t, 52, len(shuffled1), "Shuffled deck should contain 52 cards")
		assert.Equal(t, 52, len(shuffled2), "Shuffled deck should contain 52 cards")

		// 檢查洗牌後的牌組仍然包含所有原始牌
		originalCards := make(map[string]bool)
		shuffledCards1 := make(map[string]bool)
		shuffledCards2 := make(map[string]bool)

		for _, card := range deck {
			originalCards[drandshuffle.CardToString(card)] = true
		}

		for _, card := range shuffled1 {
			shuffledCards1[drandshuffle.CardToString(card)] = true
		}

		for _, card := range shuffled2 {
			shuffledCards2[drandshuffle.CardToString(card)] = true
		}

		// 確保所有原始牌都在洗牌後的牌組中
		for cardStr := range originalCards {
			assert.True(t, shuffledCards1[cardStr], "Card %s should be in shuffled deck 1", cardStr)
			assert.True(t, shuffledCards2[cardStr], "Card %s should be in shuffled deck 2", cardStr)
		}
	})
}
