package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

// Vacancy struct
type Vacancy struct {
	Title      string
	Salary     string
	Experience string
	WorkHours  string
	KeySkills  []string
	Address    string
}

func main() {
	// Instantiate default collector
	simpledCollector := colly.NewCollector(
		colly.AllowedDomains("spb.hh.ru", "www.spb.hh.ru"),
		colly.CacheDir("./spb_hh_cache"),
	)

	domain := "https://spb.hh.ru"
	detailedCollector := simpledCollector.Clone()
	idList := make([]string, 100)
	vacancies := make(map[string]*Vacancy)
	vacanciesCounter := 0
	currentPage := 30

	vacanciesFile, _ := os.OpenFile("files/vacancies.txt", os.O_APPEND|os.O_WRONLY, 0644)
	defer vacanciesFile.Close()

	// Main Page
	simpledCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if !strings.Contains(link, "vacancy/") {
			return
		}
		if matched, _ := regexp.MatchString("^[0-9]", link[strings.Index(link, "vacancy/")+len("vacancy/"):]); !matched {
			return
		}

		vacanciesCounter++
		detailedCollector.Visit(link)
	})

	// Visit next page
	simpledCollector.OnHTML("span[class=bloko-button-group]", func(e *colly.HTMLElement) {
		nextPage := ""
		isFind := false
		e.ForEach("span", func(_ int, el *colly.HTMLElement) {
			if link := el.ChildAttr("a", "href"); link != "" {
				foundPage, _ := strconv.Atoi(link[len(link)-len(strconv.Itoa(currentPage+1)):])
				if foundPage == currentPage+1 && !isFind {
					isFind = true
					nextPage = link
				}
			}
		})

		if nextPage != "" {
			if currentPage%20 == 0 {
				resFile, _ := json.MarshalIndent(vacancies, "", " ")
				vacanciesFile.Write(resFile)
				vacancies = make(map[string]*Vacancy)
			}

			currentPage++
			nextPage = domain + nextPage
			fmt.Printf("Page № %d was parsed... \n", currentPage)

			time.Sleep(time.Second * 2)
			e.Request.Visit(nextPage)
		}
	})

	// Title
	detailedCollector.OnHTML(`h1[data-qa=vacancy-title]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		title := e.Text

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].Title = title
		} else {
			vacancies[id] = &Vacancy{Title: title}
			idList = append(idList, id)
		}
	})

	// Salary
	detailedCollector.OnHTML(`p[class=vacancy-salary]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		salary := e.Text

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].Salary = salary
		} else {
			vacancies[id] = &Vacancy{Salary: salary}
			idList = append(idList, id)
		}
	})

	// Experience
	detailedCollector.OnHTML(`span[data-qa=vacancy-experience]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		experience := e.Text

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].Experience = experience
		} else {
			vacancies[id] = &Vacancy{Experience: experience}
			idList = append(idList, id)
		}
	})

	// Work Hours
	detailedCollector.OnHTML(`span[itemprop=workHours]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		workHours := e.Text

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].WorkHours = workHours
		} else {
			vacancies[id] = &Vacancy{WorkHours: workHours}
			idList = append(idList, id)
		}
	})

	// Key SkillS
	detailedCollector.OnHTML(`span[data-qa=bloko-tag__text]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		keySkill := e.Text

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].KeySkills = append(vacancies[id].KeySkills, keySkill)
		} else {
			vacancies[id] = &Vacancy{KeySkills: []string{keySkill}}
			idList = append(idList, id)
		}
	})

	// Address
	detailedCollector.OnHTML(`span[data-qa=vacancy-view-raw-address]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		address := e.Text

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].Address = address
		} else {
			vacancies[id] = &Vacancy{Address: address}
			idList = append(idList, id)
		}
	})

	detailedCollector.OnHTML(`div[data-qa=vacancy-company]`, func(e *colly.HTMLElement) {
		id := e.Request.URL.String()
		address := e.ChildText("p")

		isFind := false
		for _, value := range idList {
			if value == id {
				isFind = true
			}
		}

		if isFind {
			vacancies[id].Address = address
		} else {
			vacancies[id] = &Vacancy{Address: address}
			idList = append(idList, id)
		}
	})

	// Go Visit
	simpledCollector.Visit("https://spb.hh.ru/catalog/Informacionnye-tehnologii-Internet-Telekom/Programmirovanie-Razrabotka/page-30")

	// Show And Save Results
	fmt.Println("total vacancies: ", vacanciesCounter)
}
