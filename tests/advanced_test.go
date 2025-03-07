package tests

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"

	"go_drand/drand_shuffle"
)

// TestStringToCardErrors 測試 StringToCard 函數的錯誤處理
func TestStringToCardErrors(t *testing.T) {
	testCases := []struct {
		name        string
		cardString  string
		expectedErr string
	}{
		{
			name:        "Empty string",
			cardString:  "",
			expectedErr: "無效的牌字符串",
		},
		{
			name:        "Too short string",
			cardString:  "黑",
			expectedErr: "無效的花色",
		},
		{
			name:        "Invalid suit",
			cardString:  "藍色A",
			expectedErr: "無效的花色",
		},
		{
			name:        "Invalid value",
			cardString:  "黑桃Z",
			expectedErr: "無效的點數",
		},
		{
			name:        "Invalid value - number out of range",
			cardString:  "黑桃15",
			expectedErr: "無效的點數",
		},
		{
			name:        "Invalid value - special character",
			cardString:  "黑桃#",
			expectedErr: "無效的點數",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := drand_shuffle.StringToCard(tc.cardString)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestShuffleDeckWithInsufficientRandomness 測試 ShuffleDeck 函數處理不足的隨機性
func TestShuffleDeckWithInsufficientRandomness(t *testing.T) {
	t.Run("Test shuffle with insufficient randomness", func(t *testing.T) {
		// 創建一個非常短的隨機數據
		randomness := []byte{1, 2, 3}

		// 創建一個標準牌組
		deck := drand_shuffle.InitializeDeck()

		// 使用不足的隨機數據洗牌
		shuffled := drand_shuffle.ShuffleDeck(deck, randomness)

		// 檢查洗牌後的牌組仍然包含所有52張牌
		assert.Equal(t, 52, len(shuffled), "Shuffled deck should contain 52 cards")

		// 檢查洗牌後的牌組仍然包含所有原始牌
		originalCards := make(map[string]bool)
		shuffledCards := make(map[string]bool)

		for _, card := range deck {
			originalCards[drand_shuffle.CardToString(card)] = true
		}

		for _, card := range shuffled {
			shuffledCards[drand_shuffle.CardToString(card)] = true
		}

		// 確保所有原始牌都在洗牌後的牌組中
		for cardStr := range originalCards {
			assert.True(t, shuffledCards[cardStr], "Card %s should be in shuffled deck", cardStr)
		}
	})
}

// TestShuffleReproducibility 測試洗牌結果的可重現性
func TestShuffleReproducibility(t *testing.T) {
	t.Run("Test shuffle reproducibility with same randomness", func(t *testing.T) {
		// 創建隨機數據
		randomness := []byte("test_randomness_for_reproducibility")

		// 創建足夠的隨機性
		hasher := sha256.New()
		hasher.Write(randomness)
		extendedRandomness := hasher.Sum(nil)

		// 創建遊戲局號
		gameSessionID := "test_game_session_id"

		// 創建足夠的隨機性，加入遊戲局號
		hasher = sha256.New()
		hasher.Write(extendedRandomness)
		hasher.Write([]byte(gameSessionID))
		finalRandomness := hasher.Sum(extendedRandomness)

		// 創建一個標準牌組
		deck := drand_shuffle.InitializeDeck()

		// 使用相同的隨機數據洗牌兩次
		shuffled1 := drand_shuffle.ShuffleDeck(deck, finalRandomness)
		shuffled2 := drand_shuffle.ShuffleDeck(deck, finalRandomness)

		// 檢查兩次洗牌結果是否相同
		for i := 0; i < len(shuffled1); i++ {
			assert.Equal(t,
				drand_shuffle.CardToString(shuffled1[i]),
				drand_shuffle.CardToString(shuffled2[i]),
				"Cards at position %d should be the same", i)
		}
	})

	t.Run("Test shuffle uniqueness with different game session IDs", func(t *testing.T) {
		// 創建隨機數據
		randomness := []byte("test_randomness_for_uniqueness")

		// 創建足夠的隨機性
		hasher := sha256.New()
		hasher.Write(randomness)
		extendedRandomness := hasher.Sum(nil)

		// 創建兩個不同的遊戲局號
		gameSessionID1 := "test_game_session_id_1"
		gameSessionID2 := "test_game_session_id_2"

		// 創建足夠的隨機性，加入不同的遊戲局號
		hasher1 := sha256.New()
		hasher1.Write(extendedRandomness)
		hasher1.Write([]byte(gameSessionID1))
		finalRandomness1 := hasher1.Sum(extendedRandomness)

		hasher2 := sha256.New()
		hasher2.Write(extendedRandomness)
		hasher2.Write([]byte(gameSessionID2))
		finalRandomness2 := hasher2.Sum(extendedRandomness)

		// 創建一個標準牌組
		deck := drand_shuffle.InitializeDeck()

		// 使用不同的遊戲局號洗牌
		shuffled1 := drand_shuffle.ShuffleDeck(deck, finalRandomness1)
		shuffled2 := drand_shuffle.ShuffleDeck(deck, finalRandomness2)

		// 檢查兩次洗牌結果是否不同
		different := false
		for i := 0; i < len(shuffled1); i++ {
			if drand_shuffle.CardToString(shuffled1[i]) != drand_shuffle.CardToString(shuffled2[i]) {
				different = true
				break
			}
		}

		assert.True(t, different, "Shuffles with different game session IDs should produce different results")
	})
}
