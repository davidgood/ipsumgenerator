package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/davidgood/ipsumgenerator/wordbank"
)

var wordCountFlag, sentenceLengthFlag int
var wordbankPath string

type Service struct {
	wordBank *wordbank.WordBank
}
type IpsumResp struct {
	Ipsum string
}

type IpsumReq struct {
	Words          int
	SentenceLength int
}

func NewService(filePath string) (*Service, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	wb, err := wordbank.New(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't create the wordbank: %s", err)
	}

	return &Service{wb}, nil
}

func init() {
	flag.IntVar(&wordCountFlag, "words", 100, "number of words to generate")
	flag.IntVar(&sentenceLengthFlag, "sentence-length", 6, "the length sentences should be")
	flag.StringVar(&wordbankPath, "wordbank", "wordBank.txt", "path to the text file containing the wordbank")
}

func (svc *Service) ipsumHandler(writer http.ResponseWriter, r *http.Request) {

	req := &IpsumReq{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(req)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}

	ipsum, err := svc.generateIpsum(req.Words, req.SentenceLength)
	if err != nil {
		fmt.Println("Error:", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	enc := json.NewEncoder(writer)
	ips := &IpsumResp{Ipsum: ipsum}

	err = enc.Encode(ips)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	svc, err := NewService(wordbankPath)
	if err != nil {
		os.Exit(1)
	}

	http.HandleFunc("/ipsum", svc.ipsumHandler)

	http.ListenAndServe(":8080", nil)
}

func (svc *Service) generateIpsum(wordCount, sentenceLength int) (string, error) {

	c := make(chan string)
	sentenceCount := 0
	for wordsLeft := wordCount; wordsLeft > 0; wordsLeft -= sentenceLength {
		sentenceCount++
		numWords := sentenceLength
		if wordsLeft < numWords {
			numWords = wordsLeft
		}
		go func() {
			c <- svc.generateSentence(svc.wordBank, numWords)
		}()
	}
	ipsum := ""
	for i := 0; i < sentenceCount; i++ {
		if ipsum != "" {
			ipsum += " "
		}
		ipsum += <-c
	}
	return ipsum, nil
}

func (svc *Service) generateSentence(wb *wordbank.WordBank, wordCount int) string {
	ipsum := ""
	for i := 0; i < wordCount; i++ {
		if ipsum != "" {
			ipsum += " "
		}
		ipsum += wb.GetWord()
	}
	return ipsum + "."
}
