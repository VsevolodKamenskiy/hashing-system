package hasher

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// helper: вычисление ожидаемых хэшей (последовательно)
func mustHashSeq(t *testing.T, in []string) []string {
	t.Helper()
	ctx := context.Background()
	out, err := HashStringsParallel(ctx, in) // допускаем переиспользовать саму реализацию
	require.NoError(t, err)
	return out
}

func TestHashStringsParallel_Empty(t *testing.T) {
	ctx := context.Background()
	out, err := HashStringsParallel(ctx, nil)
	require.NoError(t, err)
	require.Len(t, out, 0)
}

func TestHashStringsParallel_Simple(t *testing.T) {
	ctx := context.Background()
	in := []string{"a", "b", "a", "hello", "世界"}
	out, err := HashStringsParallel(ctx, in)
	require.NoError(t, err)
	require.Len(t, out, len(in))

	// Детерминизм и порядок: одинаковые входы -> одинаковые хэши, и индексы совпадают
	out2, err := HashStringsParallel(ctx, in)
	require.NoError(t, err)
	require.Equal(t, out, out2)

	// Внутренне: "a" на 0 и 2 позициях одинаков
	require.Equal(t, out[0], out[2])
}

func TestHashStringsParallel_Large(t *testing.T) {
	ctx := context.Background()
	// проверим на большом массиве, что всё сходится
	n := 5000
	in := make([]string, n)
	for i := 0; i < n; i++ {
		in[i] = string(rune('a' + (i % 26)))
	}
	out, err := HashStringsParallel(ctx, in)
	require.NoError(t, err)
	require.Len(t, out, n)

	// точечные проверки детерминизма
	require.Equal(t, out[0], out[26])   // 'a' == 'a'
	require.Equal(t, out[1], out[27])   // 'b' == 'b'
	require.NotEqual(t, out[0], out[1]) // 'a' != 'b'
}

func TestHashStringsParallel_CancelEarly(t *testing.T) {
	// Проверим корректную отмену: контекст отменяется до завершения
	// Для надёжности дадим большой вход, чтобы были активные горутины
	in := make([]string, 500000)
	for i := range in {
		in[i] = "x"
	}

	ctx, cancel := context.WithCancel(context.Background())
	// отменим быстро
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	out, err := HashStringsParallel(ctx, in)
	require.Error(t, err)
	require.Nil(t, out)
}

func TestHashStringsParallel_RespectsDeadline(t *testing.T) {
	in := make([]string, 20000)
	for i := range in {
		in[i] = "payload"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	out, err := HashStringsParallel(ctx, in)
	require.Error(t, err)
	require.Nil(t, out)
}
