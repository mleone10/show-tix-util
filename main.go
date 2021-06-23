package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Customers []Customer `json:"customers"`
}

type Customer struct {
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Id           int     `json:"id"`
	Donation     float32 `json:"donation"`
	Total        float32 `json:"total"`
	CreationDate string  `json:"creation_date"`
	TenderType   string  `json:"tender_type"`
	Tickets      []struct {
		Price float32 `json:"price"`
	} `json:"tickets"`
}

type LineItem struct {
	TransactionId int
	Customer      string
	ReceiptDate   string
	DepositTo     string
	PaymentMethod string
	Memo          string
	LineItemDate  string
	LineItem      string
	Amount        float32
}

func main() {
	eventId := flag.String("event", "", "event ID to query")
	authToken := flag.String("token", "", "API token pulled from the 'connect.sid' cookie")
	flag.Parse()

	processEvent(*eventId, *authToken)
}

func processEvent(eventId, authToken string) {
	cs := []Customer{}
	var pageIndex int

	ok := true
	for ok {
		newCustomers, err := getEventPageIndex(pageIndex, eventId, authToken)
		if err != nil {
			log.Fatalf("Failed to get transactions: %v", err)
		}

		cs = append(cs, newCustomers...)
		pageIndex++

		ok = len(newCustomers) != 0
	}

	ls := parseCustomers(cs)
	printLineItems(ls)
}

func getEventPageIndex(pageNum int, eventId, authToken string) ([]Customer, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://www.showtix4u.com/api/transactions/search", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactions request: %w", err)
	}

	req.AddCookie(&http.Cookie{
		Name:  "connect.sid",
		Value: authToken,
	})

	queryParams := req.URL.Query()
	queryParams.Add("event_id", eventId)
	queryParams.Add("page", strconv.Itoa(pageNum+1))
	req.URL.RawQuery = queryParams.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call transactions api: %w", err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-success status code from transactions API: %v", res.Status)
	}

	defer res.Body.Close()

	transactions := Response{}
	json.NewDecoder(res.Body).Decode(&transactions)
	return transactions.Customers, nil
}

func parseCustomers(cs []Customer) []LineItem {
	ls := []LineItem{}

	for _, c := range cs {
		ls = append(ls, parseCustomer(c)...)
	}

	return ls
}

func parseCustomer(c Customer) []LineItem {
	ls := []LineItem{}

	for _, t := range c.Transactions {
		ls = append(ls, parseTransaction(c, t)...)
	}

	return ls
}

func parseTransaction(c Customer, t Transaction) []LineItem {
	ls := []LineItem{}
	date := formatDateString(t.CreationDate)

	if t.Donation != 0 {
		l := LineItem{
			TransactionId: t.Id,
			Customer:      fmt.Sprintf("%v, %v", strings.TrimSpace(c.LastName), strings.TrimSpace(c.FirstName)),
			ReceiptDate:   date,
			DepositTo:     "200.100 Undeposited Funds",
			PaymentMethod: t.TenderType,
			Memo:          fmt.Sprintf("Order: %v", t.Id),
			LineItemDate:  date,
			LineItem:      "Contributed Income:Unrestricted Contributions",
			Amount:        t.Donation,
		}
		ls = append(ls, l)

		l.LineItem = "Credit Card Fees"
		l.Amount = t.Donation * 0.035 * -1
		ls = append(ls, l)
	}

	for _, ticket := range t.Tickets {
		ls = append(ls, LineItem{
			TransactionId: t.Id,
			Customer:      fmt.Sprintf("%v, %v", strings.TrimSpace(c.LastName), strings.TrimSpace(c.FirstName)),
			ReceiptDate:   date,
			DepositTo:     "200.100 Undeposited Funds",
			PaymentMethod: t.TenderType,
			Memo:          fmt.Sprintf("Order: %v", t.Id),
			LineItemDate:  date,
			LineItem:      "Program Income:BO Income",
			Amount:        ticket.Price,
		})
		if ticket.Price == 0 {
			ls = append(ls, LineItem{
				TransactionId: t.Id,
				Customer:      fmt.Sprintf("%v, %v", strings.TrimSpace(c.LastName), strings.TrimSpace(c.FirstName)),
				ReceiptDate:   date,
				DepositTo:     "200.100 Undeposited Funds",
				PaymentMethod: t.TenderType,
				Memo:          fmt.Sprintf("Order: %v", t.Id),
				LineItemDate:  date,
				LineItem:      "Ticketing Fees",
				Amount:        -1.5,
			})
		}
	}

	return ls
}

func formatDateString(date string) string {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", date)
	if err != nil {
		log.Fatalf("Failed to parse date string: %v", err)
	}

	return t.Format("2006-01-02")

}

func printLineItems(ls []LineItem) {
	w := csv.NewWriter(os.Stdout)

	w.Write([]string{"Sales Receipt No", "Customer", "Sales Receipt Date", "Deposit To", "Payment Method", "Memo", "Line Item Service Date", "Line Item", "Line Item Amount"})

	for _, l := range ls {
		w.Write([]string{fmt.Sprint(l.TransactionId), l.Customer, l.ReceiptDate, l.DepositTo, l.PaymentMethod, l.Memo, l.LineItemDate, l.LineItem, fmt.Sprintf("%.2f", l.Amount)})
	}

	w.Flush()
}
