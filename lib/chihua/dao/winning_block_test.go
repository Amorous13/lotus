package dao

import (
	"fmt"
	"testing"
)

func TestWinningBlock(t *testing.T) {
	wB := NewWinningBlockDao()

	//wB.Add(10, 10, 100, []byte("1234"), "1234",[]byte("1234"))

	cid1, _ := wB.GetCIDByHeight(77398)
	fmt.Printf("%v\n", string(cid1))
	cid2, _ := wB.GetCIDByParentHeight(77397)
	fmt.Printf("%v\n", string(cid2))
	cid3, _ := wB.GetCIDByParentsTS("{bafy2bzaceakzfpn6r3ogivjq2lkfviw7t3u3mgs5z63bqlxhuvudotgspcb2g,bafy2bzacea7z6nte3hgvhliy3qrtiyrn3wfgnmbutsifqa3q6of6uvlpkpdzo,bafy2bzaceak6v3wqusobwhs2vckvuhvbbyb6wct3ozurswcxjpyzbmfronbsc,bafy2bzaced7cqoq2kcju7t5ojskhjihq4s5rsel6ehmskmp5kzbk3rmegjxks,bafy2bzaceb4n3xtqsnbcv43765dnti7krtch2qgp2ftzoww4hgnvwfiaicvtk}")
	fmt.Printf("%v\n", string(cid3))

	cid1b, err := wB.HasByHeight(77398)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("%v\n", cid1b)
	cid2b, _ := wB.HasCIDByParentHeight(77397)
	fmt.Printf("%v\n", cid2b)
	cid3b, _ := wB.HasCIDByParentsTS("{bafy2bzaceakzfpn6r3ogivjq2lkfviw7t3u3mgs5z63bqlxhuvudotgspcb2g,bafy2bzacea7z6nte3hgvhliy3qrtiyrn3wfgnmbutsifqa3q6of6uvlpkpdzo,bafy2bzaceak6v3wqusobwhs2vckvuhvbbyb6wct3ozurswcxjpyzbmfronbsc,bafy2bzaced7cqoq2kcju7t5ojskhjihq4s5rsel6ehmskmp5kzbk3rmegjxks,bafy2bzaceb4n3xtqsnbcv43765dnti7krtch2qgp2ftzoww4hgnvwfiaicvtk}")
	fmt.Printf("%v\n", cid3b)
}
