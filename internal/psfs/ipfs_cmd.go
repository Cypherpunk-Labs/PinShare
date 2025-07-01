package psfs

import (
	"fmt"
	"os/exec"
	"strings"
)

// execute this cmd ipfs add  --cid-version 1 --raw-leaves gt256kb.txt -Q
func AddFileIPFS(path string) string {
	out, err := exec.Command("ipfs", "add", "--cid-version", "1", "--raw-leaves", path, "-Q").Output()
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(strings.TrimSpace(string(out)))
	return strings.TrimSpace(string(out))
}

func GetFileIPFS(cid string, filepath string) {
	// ipfs get bafkreib566otjk54vgjqrz44xfcgdqmjgwbgatligkned7kl5qmilzvnwq  -o test.pdf
	out, err := exec.Command("ipfs", "get", cid, "-o", filepath, "--progress=false").Output()
	if err != nil {
		fmt.Println(err)
	}
	if out != nil {
	}
}

func PinFileIPFS(cid string) {
	// ipfs pin add <ipfs-path>...
	out, err := exec.Command("ipfs", "pin", "add", cid).Output()
	if err != nil {
		fmt.Println(err)
	}
	if out != nil {
	}
}

func UnpinFileIPFS(cid string) {
	// ipfs pin rm <ipfs-path>...
	out, err := exec.Command("ipfs", "pin", "rm", cid).Output()
	if err != nil {
		fmt.Println(err)
	}
	if out != nil {
	}
}
