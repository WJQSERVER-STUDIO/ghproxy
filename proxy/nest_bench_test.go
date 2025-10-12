package proxy

import (
	"ghproxy/config"
	"io"
	"strings"
	"testing"
)

const benchmarkInput = `
Some text here.
Link to be replaced: http://github.com/user/repo
Another link: https://google.com
And one more: http://example.com/some/path
This should not be replaced: notalink
End of text.
`

func BenchmarkProcessLinksStreaming(b *testing.B) {
	cfg := &config.Config{}
	host := "my-proxy.com"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		input := io.NopCloser(strings.NewReader(benchmarkInput))
		b.StartTimer()

		reader, _, err := processLinksStreamingInternal(input, host, cfg, nil)
		if err != nil {
			b.Fatalf("processLinksStreamingInternal failed: %v", err)
		}

		_, err = io.ReadAll(reader)
		if err != nil {
			b.Fatalf("Failed to read from processed reader: %v", err)
		}
	}
}

func BenchmarkProcessLinksBuffered(b *testing.B) {
	cfg := &config.Config{}
	host := "my-proxy.com"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		input := io.NopCloser(strings.NewReader(benchmarkInput))
		b.StartTimer()

		reader, _, err := processLinksBufferedInternal(input, host, cfg, nil)
		if err != nil {
			b.Fatalf("processLinksBufferedInternal failed: %v", err)
		}

		_, err = io.ReadAll(reader)
		if err != nil {
			b.Fatalf("Failed to read from processed reader: %v", err)
		}
	}
}
