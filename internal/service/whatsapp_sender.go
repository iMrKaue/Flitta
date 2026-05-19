package service

import (
	"fmt"
	"strings"
)

type WhatsAppSendResult struct {
	Provider  string
	To        string
	MessageID string
	Simulated bool
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

func SendWhatsAppMessage(to string, message string) (WhatsAppSendResult, error) {
	normalizedTo := NormalizeWhatsAppNumber(to)

	if normalizedTo == "" {
		return WhatsAppSendResult{}, fmt.Errorf("telefone do cliente não informado")
	}

	if strings.TrimSpace(message) == "" {
		return WhatsAppSendResult{}, fmt.Errorf("mensagem não informada")
	}

	fmt.Println("====================================")
	fmt.Println("SIMULAÇÃO DE ENVIO WHATSAPP")
	fmt.Println("Para:", normalizedTo)
	fmt.Println("Mensagem:")
	fmt.Println(message)
	fmt.Println("====================================")

	return WhatsAppSendResult{
		Provider:  "simulated",
		To:        normalizedTo,
		MessageID: "simulated-message-id",
		Simulated: true,
	}, nil
}
