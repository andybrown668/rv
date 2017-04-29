package van

import (
	"fmt"
	"testing"

	"github.com/d2r2/go-dht"
)

func TestDHT(t *testing.T) {
	tmp, h, err := dht.ReadDHTxx(dht.DHT22, 10, true)
	fmt.Println(tmp, h, err)
}

func BenchDHT(b *testing.B) {
	for n := 0; n < b.N; n++ {
		tmp, h, err := dht.ReadDHTxx(dht.DHT22, 10, true)
		fmt.Println(tmp, h, err)
	}
}
