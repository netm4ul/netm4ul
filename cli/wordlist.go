package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// This file is part of the setup for netm4ul.
// It helps download various wordlist available on the internet and put them into the wordlists/ folder.
// You might take a look at which one you want to use or add them by yourself.
// The users are expected to use this script but they could put directly their wordlist into the wordlists directory

// Side note : this file seems to be in a weird place, but that's the best I could find. (create another directory just for this one ? maybe later)

type wordlist struct {
	Name     string
	URL      string
	Type     string
	FileName string
}

var wordlists []wordlist

const (
	defaultDownloadDirectory = "./wordlists/"

	// Passwords
	// see https://wiki.skullsecurity.org/Passwords
	rockyouURL = "https://www.scrapmaker.com/data/wordlists/dictionaries/rockyou.txt"
	// https://github.com/danielmiessler/SecLists
	probablev2top12000URL = "https://github.com/danielmiessler/SecLists/blob/master/Passwords/probable-v2-top12000.txt"
	milwormdictURL        = "https://github.com/danielmiessler/SecLists/blob/master/Passwords/Cracked-Hashes/milw0rm-dictionary.txt"
	worstpassword500URL   = "https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/500-worst-passwords.txt"
	worst1millionURL      = "https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/10-million-password-list-top-1000000.txt"

	// (sub)domains
	subdomainsTop1MFirst5kURL  = "https://github.com/danielmiessler/SecLists/blob/master/Discovery/DNS/subdomains-top1mil-5000.txt"
	subdomainsTop1MFirst20kURL = "https://github.com/danielmiessler/SecLists/blob/master/Discovery/DNS/subdomains-top1mil-20000.txt"
	subdomainsTopFirst110kURL  = "https://github.com/danielmiessler/SecLists/blob/master/Discovery/DNS/subdomains-top1mil-110000.txt"
	jhaddixAllDNSToolURL       = "https://gist.githubusercontent.com/jhaddix/86a06c5dc309d08580a018c66354a056/raw/96f4e51d96b2203f19f6381c8c545b278eaa0837/all.txt"
)

func init() {
	//passwords
	rockyou := wordlist{Name: "Rockyou", URL: rockyouURL, Type: "passwords", FileName: "rockyou.txt"}
	probablev2top12000 := wordlist{Name: "Probable v2 top 12000", URL: probablev2top12000URL, Type: "passwords", FileName: "probablev2Top12000.txt"}
	milwormdict := wordlist{Name: "Milw0rm dictionary", URL: milwormdictURL, Type: "passwords", FileName: "milw0rm-dictionary.txt"}
	worstpassword500 := wordlist{Name: "Worst 500 password", URL: worstpassword500URL, Type: "passwords", FileName: "500-worst-passwords.txt"}
	worst1million := wordlist{Name: "Worst 1 Million password", URL: worst1millionURL, Type: "passwords", FileName: "10-million-password-list-top-1000000.txt"}

	wordlists = append(wordlists, rockyou)
	wordlists = append(wordlists, probablev2top12000)
	wordlists = append(wordlists, milwormdict)
	wordlists = append(wordlists, worstpassword500)
	wordlists = append(wordlists, worst1million)

	// domains
	subdomainsTop1MFirst5k := wordlist{Name: "Subdomains top 1 million first 5000", URL: subdomainsTop1MFirst5kURL, Type: "subdomains", FileName: "subdomains-top1mil-5000.txt"}
	subdomainsTop1MFirst20k := wordlist{Name: "Subdomains top 1 million first 20000", URL: subdomainsTop1MFirst20kURL, Type: "subdomains", FileName: "subdomains-top1mil-20000.txt"}
	subdomainsTopFirst110k := wordlist{Name: "Subdomains top 1 million first 110000", URL: subdomainsTopFirst110kURL, Type: "subdomains", FileName: "subdomains-top1mil-110000.txt"}
	jhaddixAllDNSTool := wordlist{Name: "Jhaddix \"all DNS tools\" extract", URL: jhaddixAllDNSToolURL, Type: "subdomains", FileName: "jhaddix-all-dns-tools.txt"}

	wordlists = append(wordlists, subdomainsTop1MFirst5k)
	wordlists = append(wordlists, subdomainsTop1MFirst20k)
	wordlists = append(wordlists, subdomainsTopFirst110k)
	wordlists = append(wordlists, jhaddixAllDNSTool)

}

func downloadWordlist(wl wordlist) error {

	if _, err := os.Stat(defaultDownloadDirectory); os.IsNotExist(err) {
		log.Warning("Could not find the wordlist directory, creating")
		os.Mkdir(defaultDownloadDirectory, os.ModePerm)
	}

	// Create the file
	out, err := os.Create(path.Join(defaultDownloadDirectory, wl.FileName))
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(wl.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// This function prints all the wordlist corresponding to a type and return the number of elements printed.
// We use this count to make the boundaries for the selection.
func printWordlistByType(t string) int {
	//We are using a external index because the "wordlists" var contains different types of wordlist (password, subdomains, usernames...)
	index := 0
	for _, wl := range wordlists {
		if wl.Type == t {
			fmt.Printf("[%d] %s\n", index, wl.Name)
			index++
		}
	}
	return index
}

func getWordlistByTypeAndIndex(t string, selectedPasswordIndex int) (wordlist, error) {
	index := 0
	for _, wl := range wordlists {
		if wl.Type == t {
			if selectedPasswordIndex == index {
				return wl, nil
			}
			index++
		}
	}
	return wordlist{}, errors.New("Could not find the provided wordlist (type : " + t + ", index : " + strconv.Itoa(selectedPasswordIndex) + ")")
}
