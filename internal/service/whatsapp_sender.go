package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type WhatsAppSendResult struct {
	Provider  string `json:"provider"`
	To        string `json:"to"`
	MessageID string `json:"message_id"`
	Simulated bool   `json:"simulated"`
}

type twilioMessageResponse struct {
	SID     string `json:"sid"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NormalizeWhatsAppNumber(phone string) string {
	phone = strings.TrimSpace(phone)

	if phone == "" {
		return ""
	}

	if strings.HasPrefix(phone, "whatsapp:") {
		return phone
	}

	return "whatsapp:" + phone
}

func whatsappProvider() string {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("WHATSAPP_PROVIDER")))

	if provider == "" {
		return "simulated"
	}

	return provider
}

// SendAppointmentReminder é usado pelos lembretes.
// Em modo simulado, mostra a mensagem personalizada no terminal.
// Em modo Twilio Sandbox, usa o template aprovado de lembrete.
func SendAppointmentReminder(
	to string,
	simulatedMessage string,
	appointmentDate string,
	appointmentTime string,
) (WhatsAppSendResult, error) {
	normalizedTo := NormalizeWhatsAppNumber(to)

	if normalizedTo == "" {
		return WhatsAppSendResult{}, fmt.Errorf("telefone do cliente não informado")
	}

	if whatsappProvider() == "twilio" {
		return sendTwilioAppointmentReminder(normalizedTo, appointmentDate, appointmentTime)
	}

	return simulateWhatsAppMessage(normalizedTo, simulatedMessage)
}

// SendWhatsAppMessage permanece disponível para mensagens livres.
// No Sandbox, use mensagens livres apenas quando existir janela de atendimento aberta.
func SendWhatsAppMessage(to string, message string) (WhatsAppSendResult, error) {
	normalizedTo := NormalizeWhatsAppNumber(to)

	if normalizedTo == "" {
		return WhatsAppSendResult{}, fmt.Errorf("telefone do cliente não informado")
	}

	if strings.TrimSpace(message) == "" {
		return WhatsAppSendResult{}, fmt.Errorf("mensagem não informada")
	}

	if whatsappProvider() == "twilio" {
		return sendTwilioFreeformMessage(normalizedTo, message)
	}

	return simulateWhatsAppMessage(normalizedTo, message)
}

func simulateWhatsAppMessage(to string, message string) (WhatsAppSendResult, error) {
	if strings.TrimSpace(message) == "" {
		return WhatsAppSendResult{}, fmt.Errorf("mensagem não informada")
	}

	fmt.Println("====================================")
	fmt.Println("SIMULAÇÃO DE ENVIO WHATSAPP")
	fmt.Println("Para:", to)
	fmt.Println("Mensagem:")
	fmt.Println(message)
	fmt.Println("====================================")

	return WhatsAppSendResult{
		Provider:  "simulated",
		To:        to,
		MessageID: "simulated-message-id",
		Simulated: true,
	}, nil
}

func sendTwilioAppointmentReminder(
	to string,
	appointmentDate string,
	appointmentTime string,
) (WhatsAppSendResult, error) {
	contentSID := strings.TrimSpace(os.Getenv("TWILIO_APPOINTMENT_TEMPLATE_SID"))
	if contentSID == "" {
		return WhatsAppSendResult{}, fmt.Errorf("TWILIO_APPOINTMENT_TEMPLATE_SID não configurado")
	}

	contentVariables, err := json.Marshal(map[string]string{
		"1": appointmentDate,
		"2": appointmentTime,
	})
	if err != nil {
		return WhatsAppSendResult{}, fmt.Errorf("erro ao montar variáveis do template: %w", err)
	}

	form := url.Values{}
	form.Set("From", NormalizeWhatsAppNumber(os.Getenv("TWILIO_WHATSAPP_FROM")))
	form.Set("To", to)
	form.Set("ContentSid", contentSID)
	form.Set("ContentVariables", string(contentVariables))

	return sendTwilioRequest(form, to)
}

func sendTwilioFreeformMessage(to string, message string) (WhatsAppSendResult, error) {
	form := url.Values{}
	form.Set("From", NormalizeWhatsAppNumber(os.Getenv("TWILIO_WHATSAPP_FROM")))
	form.Set("To", to)
	form.Set("Body", message)

	return sendTwilioRequest(form, to)
}

func sendTwilioRequest(form url.Values, to string) (WhatsAppSendResult, error) {
	accountSID := strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID"))
	authToken := strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN"))
	from := strings.TrimSpace(form.Get("From"))

	if accountSID == "" {
		return WhatsAppSendResult{}, fmt.Errorf("TWILIO_ACCOUNT_SID não configurado")
	}

	if authToken == "" {
		return WhatsAppSendResult{}, fmt.Errorf("TWILIO_AUTH_TOKEN não configurado")
	}

	if from == "" || from == "whatsapp:" {
		return WhatsAppSendResult{}, fmt.Errorf("TWILIO_WHATSAPP_FROM não configurado")
	}

	endpoint := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		accountSID,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return WhatsAppSendResult{}, fmt.Errorf("erro ao criar requisição Twilio: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(accountSID, authToken)

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return WhatsAppSendResult{}, fmt.Errorf("erro ao conectar com Twilio: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return WhatsAppSendResult{}, fmt.Errorf("erro ao ler resposta Twilio: %w", err)
	}

	var twilioResponse twilioMessageResponse
	_ = json.Unmarshal(body, &twilioResponse)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		message := twilioResponse.Message
		if message == "" {
			message = string(body)
		}

		return WhatsAppSendResult{}, fmt.Errorf("Twilio recusou o envio: %s", message)
	}

	return WhatsAppSendResult{
		Provider:  "twilio",
		To:        to,
		MessageID: twilioResponse.SID,
		Simulated: false,
	}, nil
}
