package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Product struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Image     string  `json:"image"`
	Price     float64 `json:"price"`
	Sales     int64   `json:"sales"`
	Stock     int64   `json:"stock"`
	Favorites int64   `json:"favorites,omitempty"`
}

var (
	port         string
	favorites    bool
	salesURL     string
	stockURL     string
	favoritesURL string
)

func init() {
	port = GetEnvDefault("port", "7000")
	favorites = getBoolEnv("favorites")
	salesURL = GetEnvDefault("salesURL", "http://sales:7000")
	stockURL = GetEnvDefault("stockURL", "http://stock:7000")
	favoritesURL = GetEnvDefault("favoritesURL", "http://favorites:7000")
}

func getBoolEnv(key string) bool {
	_, ex := os.LookupEnv(key)
	return ex
}

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func main() {
	flag.Parse()
	http.HandleFunc("/products", productsController)
	fmt.Println("staring products service on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func productsController(w http.ResponseWriter, r *http.Request) {
	var ids []int64
	v := r.URL.Query()
	query := strings.Split(v.Get("ids"), ",")

	for _, q := range query {
		id, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameter ids is required"))
		return
	}

	fmt.Printf("querying products of %v\n", ids)
	headers := getForwardHeaders(r)

	products := getProduct(ids, headers)

	js, err := json.Marshal(products)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getProduct(ids []int64, headers map[string]string) []*Product {
	var products []*Product

	var idsStr []string
	for _, id := range ids {
		if p, ok := mockDB[id]; ok {
			products = append(products, &p)
			idsStr = append(idsStr, strconv.FormatInt(id, 10))
		}
	}

	if len(products) == 0 {
		return products
	}

	var waitGroup sync.WaitGroup
	query := strings.Join(idsStr, ",")

	waitGroup.Add(1)
	go func() {
		fmt.Println("call sales service")
		result := make(map[int64]int64)
		err := getJson(fmt.Sprintf(salesURL+"/sales?ids=%s", query), &result, headers)
		fmt.Println(err)
		for _, p := range products {
			if v, ok := result[p.ID]; ok {
				p.Sales = v
			}
		}
		waitGroup.Done()
	}()

	waitGroup.Add(1)
	go func() {
		fmt.Println("call stock service")
		result := make(map[int64]int64)
		err := getJson(fmt.Sprintf(stockURL+"/stock?ids=%s", query), &result, headers)
		fmt.Println(err)
		for _, p := range products {
			if v, ok := result[p.ID]; ok {
				p.Stock = v
			}
		}
		waitGroup.Done()
	}()

	if favorites {
		waitGroup.Add(1)
		go func() {
			fmt.Println("call favorites service")
			result := make(map[int64]int64)
			err := getJson(fmt.Sprintf(favoritesURL+"/favorites?ids=%s", query), &result, headers)
			if err != nil {
				fmt.Println(err)
			}
			for _, p := range products {
				fmt.Println(result[p.ID])
				if v, ok := result[p.ID]; ok {
					p.Favorites = v
				}
			}
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
	return products
}

func getJson(url string, target interface{}, headers map[string]string) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("GET", url, nil)

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	if err != nil {
		return err
	}
	response, err := client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	return json.NewDecoder(response.Body).Decode(target)
}

func getForwardHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	forwardHeaders := []string{
		"user",
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-ot-span-context",
	}

	for _, h := range forwardHeaders {
		if v := r.Header.Get(h); v != "" {
			headers[h] = v
		}
	}

	return headers
}

var mockDB = map[int64]Product{
	1: {
		ID:    1,
		Title: "Kubernetes",
		Image: "1",
		Price: 8,
	},
	2: {
		ID:    2,
		Title: "Docker",
		Image: "2",
		Price: 8,
	},
	3: {
		ID:    3,
		Title: "Istio",
		Image: "3",
		Price: 15,
	},
	4: {
		ID:    4,
		Title: "Prometheus",
		Image: "4",
		Price: 35,
	},
	5: {
		ID:    5,
		Title: "Jaeger",
		Image: "5",
		Price: 0.5,
	},
	6: {
		ID:    6,
		Title: "RabbitMQ",
		Image: "6",
		Price: 15,
	},
	7: {
		ID:    7,
		Title: "Golang",
		Image: "7",
		Price: 35,
	},
	8: {
		ID:    8,
		Title: "Python",
		Image: "8",
		Price: 15,
	},
	9: {
		ID:    9,
		Title: "Ruby",
		Image: "9",
		Price: 8,
	},
	10: {
		ID:    10,
		Title: "MongoDB",
		Image: "10",
		Price: 15,
	},
	11: {
		ID:    11,
		Title: "Envoy",
		Image: "11",
		Price: 15,
	},
	12: {
		ID:    12,
		Title: "Java",
		Image: "12",
		Price: 0.5,
	},
	13: {
		ID:    13,
		Title: "NodeJS",
		Image: "13",
		Price: 35,
	},
	14: {
		ID:    14,
		Title: "Calico",
		Image: "14",
		Price: 8,
	},
	15: {
		ID:    15,
		Title: "Redis",
		Image: "15",
		Price: 19,
	},
}
