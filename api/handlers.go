package api

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"

    "TransactionSystem/internal/models"

    "github.com/gorilla/mux"
)

func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Balance float64 `json:"balance"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        log.Printf("CreateWallet: failed to parse request body: %v", err)
        return
    }

    address, err := h.walletService.CreateWallet(r.Context(), req.Balance)
    if err != nil {
        http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
        log.Printf("CreateWallet: failed to create wallet: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"address": address})
}


func (h *Handler) RemoveWallet(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    address, ok := vars["address"]
    if !ok {
        http.Error(w, "Address is required", http.StatusBadRequest)
        return
    }

    err := h.walletService.RemoveWallet(r.Context(), address)
    if err != nil {
        http.Error(w, "Failed to remove wallet", http.StatusInternalServerError)
        log.Printf("RemoveWallet: failed to remove wallet: %v", err)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    address, ok := vars["address"]
    if !ok {
        http.Error(w, "Address is required", http.StatusBadRequest)
        return
    }

    var wallet *models.Wallet

    wallet, err := h.walletService.GetWallet(r.Context(), address)
    if err != nil {
        http.Error(w, "Failed to get wallet", http.StatusInternalServerError)
        log.Printf("GetWalletInfo: failed to get wallet: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(wallet)
}

func (h *Handler) GetTransactionById(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(w, "ID is required", http.StatusBadRequest)
        return
    }

    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var transaction *models.Transaction

    transaction, err = h.transactionService.GetTransactionById(r.Context(), id)
    if err != nil {
        http.Error(w, "Transaction not found", http.StatusNotFound)
        log.Printf("GetTransactionById: transaction not found: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(transaction)
}

func (h *Handler) RemoveTransaction(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(w, "ID is required", http.StatusBadRequest)
        return
    }

    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    err = h.transactionService.RemoveTransaction(r.Context(), id)
    if err != nil {
        http.Error(w, "Failed to remove transaction", http.StatusInternalServerError)
        log.Printf("RemoveTransaction: failed to remove transaction: %v", err)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetTransactionByInfo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    from, ok := vars["from"]
    if !ok {
        http.Error(w, "Sender address (from) is required", http.StatusBadRequest)
        return
    }

    to, ok := vars["to"]
    if !ok {
        http.Error(w, "Receiver address (to) is required", http.StatusBadRequest)
        return
    }

    createdAtStr, ok := vars["createdAt"]
    if !ok {
        http.Error(w, "Transaction timestamp (createdAt) is required", http.StatusBadRequest)
        return
    }

    createdAt, err := time.Parse(time.RFC3339, createdAtStr)
    if err != nil {
        http.Error(w, "Invalid timestamp format. Use RFC3339 format (e.g., 2024-02-10T15:04:05Z)", http.StatusBadRequest)
        return
    }

    transaction, err := h.transactionService.GetTransactionByInfo(r.Context(), from, to, createdAt)
    if err != nil {
        http.Error(w, "Transaction not found", http.StatusNotFound)
        log.Printf("GetTransactionByInfo: transaction not found: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(transaction)
}