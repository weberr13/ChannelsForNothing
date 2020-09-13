package example1

import (
	"context"
	"testing"
	"time"
)

func Benchmark10WritersAsyncSync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 10, 1)
	b.ResetTimer()
	asyncSyncMode(ctx, b, async, names, b.N, true)
	cancel()
}
func Benchmark10WritersAsyncAsync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 10, 1)
	b.ResetTimer()
	fullAsync(ctx, b, async, names, b.N, false)
	cancel()
}
func Benchmark10WritersBurstAsync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 10, 1)
	b.ResetTimer()
	burstMode(ctx, b, async, names, b.N, false)
	cancel()
}
func Benchmark10WritersBurstSync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(10, 1)
	b.ResetTimer()
	burstSync(b, sync, names, b.N, true)
}
func Benchmark10WritersSyncSync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(10, 1)
	b.ResetTimer()
	syncSync(b, sync, names, b.N, true)
}
func Benchmark10WritersSyncAsync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(10, 1)

	b.ResetTimer()
	syncAsync(b, sync, names, b.N, false)
}

func Benchmark100WritersAsyncSync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 100, 1)
	b.ResetTimer()
	asyncSyncMode(ctx, b, async, names, b.N, true)
	cancel()
}
func Benchmark100WritersAsyncAsync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 100, 1)
	b.ResetTimer()
	fullAsync(ctx, b, async, names, b.N, false)
	cancel()
}

func Benchmark100WritersSyncSync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(100, 1)
	b.ResetTimer()
	syncSync(b, sync, names, b.N, true)
}
func Benchmark100WritersSyncAsync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(100, 1)

	b.ResetTimer()
	syncAsync(b, sync, names, b.N, false)
}
func Benchmark100WritersBurstAsync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 100, 1)
	b.ResetTimer()
	burstMode(ctx, b, async, names, b.N, false)
	cancel()
}
func Benchmark100WritersBurstSync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(100, 1)
	b.ResetTimer()
	burstSync(b, sync, names, b.N, true)
}
func Benchmark500WritersAsyncSync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 500, 1)
	b.ResetTimer()
	asyncSyncMode(ctx, b, async, names, b.N, true)
	cancel()
}

func Benchmark500WritersAsyncAsync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 500, 1)
	b.ResetTimer()
	fullAsync(ctx, b, async, names, b.N, false)
	cancel()
}

func Benchmark500WritersSyncSync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(500, 1)
	b.ResetTimer()
	syncSync(b, sync, names, b.N, true)
}

func Benchmark500WritersSyncAsync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(500, 1)

	b.ResetTimer()
	syncAsync(b, sync, names, b.N, false)
}

func Benchmark500WritersBurstAsync(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async, names := getWriter(ctx, b, 500, 1)
	b.ResetTimer()
	burstMode(ctx, b, async, names, b.N, false)
	cancel()
}
func Benchmark500WritersBurstSync(b *testing.B) {
	b.ReportAllocs()
	sync, names := getSyncWriter(500, 1)
	b.ResetTimer()
	burstSync(b, sync, names, b.N, true)
}

// // ratio tests

// func Benchmark10WritersAsyncSync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
// 	async, names := getWriter(ctx, b, 10, 10)
// 	b.ResetTimer()
// 	asyncSyncMode(ctx, b, async, names, b.N, true)
// 	cancel()
// }
// func Benchmark10WritersAsyncAsync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
// 	async, names := getWriter(ctx, b, 10, 10)
// 	b.ResetTimer()
// 	fullAsync(ctx, b, async, names, b.N, false)
// 	cancel()
// }
// func Benchmark10WritersSyncSync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	sync, names := getSyncWriter(10, 10)
// 	b.ResetTimer()
// 	syncSync(b, sync, names, b.N, true)
// }
// func Benchmark10WritersSyncAsync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	sync, names := getSyncWriter(10, 10)

// 	b.ResetTimer()
// 	syncAsync(b, sync, names, b.N, false)
// }
// func Benchmark100WritersAsyncSync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
// 	async, names := getWriter(ctx, b, 100, 10)
// 	b.ResetTimer()
// 	asyncSyncMode(ctx, b, async, names, b.N, true)
// 	cancel()
// }
// func Benchmark100WritersAsyncAsync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
// 	async, names := getWriter(ctx, b, 100, 10)
// 	b.ResetTimer()
// 	fullAsync(ctx, b, async, names, b.N, false)
// 	cancel()
// }

// func Benchmark100WritersSyncSync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	sync, names := getSyncWriter(100, 10)
// 	b.ResetTimer()
// 	syncSync(b, sync, names, b.N, true)
// }
// func Benchmark100WritersSyncAsync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	sync, names := getSyncWriter(100, 10)

// 	b.ResetTimer()
// 	syncAsync(b, sync, names, b.N, false)
// }
// func Benchmark500WritersAsyncSync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
// 	async, names := getWriter(ctx, b, 500, 10)
// 	b.ResetTimer()
// 	asyncSyncMode(ctx, b, async, names, b.N, true)
// 	cancel()
// }

// func Benchmark500WritersAsyncAsync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
// 	async, names := getWriter(ctx, b, 500, 10)
// 	b.ResetTimer()
// 	fullAsync(ctx, b, async, names, b.N, false)
// 	cancel()
// }

// func Benchmark500WritersSyncSync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	sync, names := getSyncWriter(500, 10)
// 	b.ResetTimer()
// 	syncSync(b, sync, names, b.N, true)
// }

// func Benchmark500WritersSyncAsync10Prep(b *testing.B) {
// 	b.ReportAllocs()
// 	sync, names := getSyncWriter(500, 10)

// 	b.ResetTimer()
// 	syncAsync(b, sync, names, b.N, false)
// }
