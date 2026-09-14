package goip

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GoIPClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewIPClient(ip, username, password string) *GoIPClient {
	return &GoIPClient{
		baseURL:  fmt.Sprintf("http://%s/goip/goip", ip),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type InboxResponse struct {
	Status   string        `json:"status"`
	Messages []IncomingSMS `json:"messages"`
}
type IncomingSMS struct {
	Slot   int    `json:"slot"`   // слот, на который пришло SMS
	Sender string `json:"sender"` // номер отправителя
	Time   string `json:"time"`   // время получения
	Text   string `json:"text"`   // текст сообщения
}
type SlotStatus struct {
	Slot     int    `json:"slot"`
	Number   string `json:"number"`
	Signal   int    `json:"signal"`
	Operator string `json:"operator"`
	Status   string `json:"status"`
}
type StatusResponse struct {
	Status string       `json:"status"`
	Slots  []SlotStatus `json:"slots"`
}

func (i *GoIPClient) GetInfo() (*StatusResponse, error) {
	params := url.Values{}
	params.Add("username", i.username)
	params.Add("password", i.password)
	params.Add("cmd", "1")
	resp, err := i.client.Get(i.baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}
	var result StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}
	return &result, nil
}
func (i *GoIPClient) SendSMS(phone, slot, text string) error {
	params := url.Values{}
	params.Add("username", i.username)
	params.Add("password", i.password)
	params.Add("cmd", "2")
	params.Add("slot", slot)
	params.Add("phone", phone)
	params.Add("text", text)
	resp, err := i.client.Get(i.baseURL + "?" + params.Encode())
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (i *GoIPClient) GetInbox() (*InboxResponse, error) { // банк сообщений
	params := url.Values{}
	params.Add("username", i.username)
	params.Add("password", i.password)
	params.Add("cmd", "3")
	resp, err := i.client.Get(i.baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}
	var result InboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}
	return &result, nil
}
