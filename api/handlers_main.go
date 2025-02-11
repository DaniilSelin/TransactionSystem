package api

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"

    "TransactionSystem/internal/service"

    "github.com/gorilla/mux"
)

type Handler struct {
    transactionService *service.TransactionService
    walletService      *service.WalletService
}

func NewHandler(ts *service.TransactionService, ws *service.WalletService) *Handler {
    return &Handler{
        transactionService: ts,
        walletService:      ws,
    }
}

func (h *Handler) SendMoney(w http.ResponseWriter, r *http.Request) {
    var req struct {
        From   string  `json:"from"`
        To     string  `json:"to"`
        Amount float64 `json:"amount"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        log.Printf("SendMoney: failed to decode request: %v", err)
        return
    }

    err := h.transactionService.SendMoney(r.Context(), req.From, req.To, req.Amount)
    if err != nil {
        http.Error(w, "Transaction failed", http.StatusBadRequest)
        log.Printf("SendMoney: transaction failed: %v", err)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetLastTransactions(w http.ResponseWriter, r *http.Request) {
    countStr := r.URL.Query().Get("count")
    count, err := strconv.Atoi(countStr)
    if err != nil || count <= 0 {
        http.Error(w, "Invalid count parameter", http.StatusBadRequest)
        log.Printf("GetLastTransactions: invalid count parameter: %v", err)
        return
    }

    transactions, err := h.transactionService.GetLastTransactions(r.Context(), count)
    if err != nil {
        http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
        log.Printf("GetLastTransactions: failed to fetch transactions: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(transactions)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    address, ok := vars["address"]
    if !ok {
        http.Error(w, "Wallet address is required", http.StatusBadRequest)
        return
    }

    balance, err := h.walletService.GetBalance(r.Context(), address)
    if err != nil {
        http.Error(w, "Failed to fetch balance", http.StatusInternalServerError)
        log.Printf("GetBalance: failed to fetch balance: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}