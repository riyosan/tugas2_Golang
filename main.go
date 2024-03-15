package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func connectDB() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/orders_by")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MySQL database")
}

type Item struct {
	ItemID      int
	ItemCode    string
	Description string
	Quantity    int
	OrderID     int
}

type Order struct {
	OrderID      int
	CustomerName string
	OrderedAt    string
	Items        []Item
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var newOrder Order
	err := json.NewDecoder(r.Body).Decode(&newOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert order into database
	result, err := db.Exec("INSERT INTO orders (customer_name, ordered_at) VALUES (?, ?)",
		newOrder.CustomerName, newOrder.OrderedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderID, _ := result.LastInsertId()

	// Insert items into database
	for _, item := range newOrder.Items {
		_, err = db.Exec("INSERT INTO items (item_code, description, quantity, order_id) VALUES (?, ?, ?, ?)",
			item.ItemCode, item.Description, item.Quantity, orderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Order created successfully")
}

// Fungsi untuk mengambil data order
func getOrderData(orderID int) (*Order, error) {
	// Buka koneksi database
	connectDB()

	// Lakukan query ke database
	rows, err := db.Query("SELECT * FROM orders WHERE order_id = ?", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Lakukan iterasi baris hasil query
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.OrderID, &order.CustomerName, &order.OrderedAt); err != nil {
			return nil, err
		}
		return &order, nil
	}

	return nil, errors.New("Order not found")
}

// Handler untuk mendapatkan semua order
func getOrders(w http.ResponseWriter, r *http.Request) {
	// Buka koneksi database
	connectDB()

	// Lakukan query ke database untuk mendapatkan semua order
	rows, err := db.Query("SELECT * FROM orders")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Buat slice untuk menyimpan semua order
	var orders []Order

	// Lakukan iterasi baris hasil query
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.OrderID, &order.CustomerName, &order.OrderedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Tambahkan order ke slice
		orders = append(orders, order)
	}

	// Konversi slice order menjadi format JSON
	jsonData, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response header dan kirim respons
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Handler untuk memperbarui order
func updateOrder(w http.ResponseWriter, r *http.Request) {
	// Buka koneksi database
	connectDB()

	// Ambil orderId dari URL
	orderId := r.URL.Query().Get("orderId")

	// Parse request body
	var updatedOrder Order
	err := json.NewDecoder(r.Body).Decode(&updatedOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lakukan query ke database untuk memperbarui order
	_, err = db.Exec("UPDATE orders SET customer_name = ?, ordered_at = ? WHERE order_id = ?",
		updatedOrder.CustomerName, updatedOrder.OrderedAt, orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim respons sukses
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Order updated successfully")
}

// Handler untuk menghapus order
func deleteOrder(w http.ResponseWriter, r *http.Request) {
	// Buka koneksi database
	connectDB()

	// Ambil orderId dari URL
	orderId := r.URL.Query().Get("orderId")

	// Lakukan query ke database untuk menghapus order
	_, err := db.Exec("DELETE FROM orders WHERE order_id = ?", orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim respons sukses
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Order deleted successfully")
}

func main() {
	connectDB()

	// Endpoint handlers
	http.HandleFunc("/create-order", createOrder)
	http.HandleFunc("/get-orders", getOrders)
	http.HandleFunc("/update-order", updateOrder)
	http.HandleFunc("/delete-order", deleteOrder) // Mengubah path menjadi /delete-order untuk menghindari konflik

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
