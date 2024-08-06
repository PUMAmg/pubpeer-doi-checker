package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Pubpeer struct {
	Publications []struct {
		ID                int    `json:"id"`
		Title             string `json:"title"`
		Abstract          string `json:"abstract"`
		PubpeerID         string `json:"pubpeer_id"`
		PublishedAt       string `json:"published_at"`
		LinkWithHash      string `json:"link_with_hash"`
		Created           string `json:"created"`
		CreatedDiff       string `json:"created_diff"`
		Updated           string `json:"updated"`
		UpdatedDiff       string `json:"updated_diff"`
		LastCommented     string `json:"last_commented"`
		LastCommentedDiff string `json:"last_commented_diff"`
		CommentsTotal     int    `json:"comments_total"`
		HasAuthorResponse bool   `json:"has_author_response"`
		JournalHasTeam    bool   `json:"journal_has_team"`
		LastCommentID     int    `json:"last_comment_id"`
		AffiliationList   string `json:"affiliation_list"`
		Authors           struct {
			Data []struct {
				ID           int           `json:"id"`
				FirstName    string        `json:"first_name"`
				LastName     string        `json:"last_name"`
				DisplayName  string        `json:"display_name"`
				Email        interface{}   `json:"email"`
				Affiliations []interface{} `json:"affiliations"`
			} `json:"data"`
		} `json:"authors"`
		Journals struct {
			Data []struct {
				ID    int    `json:"id"`
				Title string `json:"title"`
				Issn  string `json:"issn"`
			} `json:"data"`
		} `json:"journals"`
		Updates struct {
			Data []interface{} `json:"data"`
		} `json:"updates"`
	} `json:"publications"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

func main() {
	fmt.Println("Список файлов:", os.Args[1:])
	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		result, err := os.Create(strings.TrimSuffix(filename, ".txt") + " result.txt")
		if err != nil {
			fmt.Println("Error create file:", err)
			return
		}
		buf := bytes.NewBufferString("")
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			doi := ReadDOI(scanner)
			time.Sleep(time.Second * 2)
			str, err := CurlPubPeer(doi)
			if err != nil {
				fmt.Println("Error get info from pubpeer:", err, "doi:", doi)
				buf.Write([]byte(fmt.Sprintf("%s - error\n", doi)))
				continue
			}
			buf.Write([]byte(str))
		}
		err = WriteFile(buf, result)
		if err != nil {
			fmt.Println("Error write file:", err)
			return
		}
		file.Close()
		result.Close()
	}
}

func ReadDOI(scanner *bufio.Scanner) string {
	return scanner.Text()
}

func CurlPubPeer(doi string) (string, error) {
	var pubs Pubpeer
	url := fmt.Sprintf("https://pubpeer.com/api/search/?q=%s", doi)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error get req:", err)
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error read body:", err)
		return "", err
	}
	err = json.Unmarshal(body, &pubs)
	if err != nil {
		fmt.Println("Error unmarshal body:", err)
		return "", err
	}
	if len(pubs.Publications) > 0 {
		return fmt.Sprintf("%s&да&%d&%s\n", strings.TrimSuffix(doi, "."), pubs.Publications[0].CommentsTotal, pubs.Publications[0].Title), nil
	}
	return fmt.Sprintf("%s&нет\n", doi), nil
}

func WriteFile(buf *bytes.Buffer, file *os.File) error {
	_, err := file.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
