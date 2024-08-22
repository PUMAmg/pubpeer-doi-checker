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

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/gosuri/uilive"
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
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	for {
		fmt.Fprintf(writer, "Что проверяем? Введите цифру команды:\n1.Проверка на PubPeer по DOI\n2.Проверка SJR\nНажмите q для выхода\r\n")
		char, _, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		switch char {
		case '1':
			for {
				fmt.Fprintf(writer, "Вы выбрали проверку на PubPeer по DOI, для этого в папку с этой программой поместите файлы с DOI в формате txt. Нажмите 1 для продолжения\r\n")
				char, _, err := keyboard.GetKey()
				if err != nil {
					panic(err)
				}
				if char == '1' {
					PubPeerCheck(writer)
					break
				}
			}
		case '2':
			for {
				fmt.Fprintf(writer, "Вы выбрали проверку SJR, для этого в папку с программой поместите файлы со списком ISSN или названий журналов. Нажмите 2 для продолжения\r\n")
				char, _, err := keyboard.GetKey()
				if err != nil {
					panic(err)
				}
				if char == '2' {
					fmt.Fprintf(writer, "Тут пока ничего нет, возврат в меню.\r\n\n")
					break
				}
			}
		case 'q', 'Q', 'й', 'Й':
			fmt.Fprintf(writer, "Выход\r\n")
			return
		}
	}
}

func PubPeerCheck(writer *uilive.Writer) {
	files := GetDOIFiles()
	fmt.Fprintf(writer, "Список файлов: %s\r\n", files)

	for _, filename := range files {
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
			doi := scanner.Text()
			time.Sleep(time.Second * 2)
			pubs, err := CurlPubPeer(doi)
			if err != nil {
				fmt.Println("Error get info from pubpeer:", err, "doi:", doi)
				buf.Write([]byte(fmt.Sprintf("%s - error\n", doi)))
				continue
			}
			var str string
			if len(pubs.Publications) > 0 {
				str = fmt.Sprintf("%s&да&%d&%s\n", strings.TrimSuffix(doi, "."), pubs.Publications[0].CommentsTotal, pubs.Publications[0].Title)
			} else {
				str = fmt.Sprintf("%s&нет\n", doi)
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
	green := color.New(color.FgGreen)
	green.Fprintf(writer, "Проверка завершена!\r\n\n")
}

func GetDOIFiles() []string {
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error read dir:", err)
		return nil
	}
	var DOIFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			if strings.HasSuffix(file.Name(), "result.txt") {
				continue
			}
			DOIFiles = append(DOIFiles, file.Name())
		}
	}
	return DOIFiles
}

func CurlPubPeer(doi string) (Pubpeer, error) {
	var pubs Pubpeer
	url := fmt.Sprintf("https://pubpeer.com/api/search/?q=%s", doi)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error get req:", err)
		return pubs, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error read body:", err)
		return pubs, err
	}
	err = json.Unmarshal(body, &pubs)
	if err != nil {
		fmt.Println("Error unmarshal body:", err)
		return pubs, err
	}
	if pubs.Publications == nil {
		return pubs, fmt.Errorf("ошибка получения данных")
	}
	return pubs, nil
}

func WriteFile(buf *bytes.Buffer, file *os.File) error {
	_, err := file.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
