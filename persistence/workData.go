package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v4"
	"log"
	"time"
)

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Items struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Items   `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

const (
	//sql для получение всех заказов с items
	sqlTakeAllDeliveries string = `SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
       o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
       p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
       i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
       FROM deliveries d
       JOIN order_deliveries od on d.id = od.delivery_id
       Join orders o on od.order_uid = o.order_uid
       JOIN payments p on p.transaction = o.order_uid
       Left Join items i on o.track_number = i.track_number;`

	sqlInsertPayment              = "INSERT INTO payments VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);"
	sqlInsertOrder         string = "INSERT INTO orders(order_uid, track_number, entry, locale, internal_signature,customer_id, delivery_service, shardkey, sm_id, date_created,oof_shard) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)"
	sqlInsertItem          string = "INSERT INTO items(chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"
	sqlInsertDelivery      string = "INSERT INTO deliveries (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;"
	sqlInsertOrderDelivery string = "INSERT INTO order_deliveries VALUES ($1, $2);"
)

var Cache map[string]Order

// InitCache отвечает за инициализацию кэша
func InitCache(db *pgx.Conn) error {
	// Инициализируем кэш
	Cache = make(map[string]Order)
	//Получаем все строки с заказами
	rows, err := db.Query(context.Background(), sqlTakeAllDeliveries)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	//Инициализируем строки в кэшэ
	for rows.Next() {
		var Or Order
		var It Items
		err := rows.Scan(&Or.OrderUID, &Or.TrackNumber, &Or.Entry, &Or.Locale, &Or.InternalSignature,
			&Or.CustomerID, &Or.DeliveryService, &Or.Shardkey, &Or.SmID, &Or.DateCreated, &Or.OofShard,
			&Or.Delivery.Name, &Or.Delivery.Phone, &Or.Delivery.Zip, &Or.Delivery.City, &Or.Delivery.Address, &Or.Delivery.Region, &Or.Delivery.Email,
			&Or.Payment.Transaction, &Or.Payment.RequestID, &Or.Payment.Currency, &Or.Payment.Provider, &Or.Payment.Amount, &Or.Payment.PaymentDt, &Or.Payment.Bank, &Or.Payment.DeliveryCost, &Or.Payment.GoodsTotal, &Or.Payment.CustomFee,
			&It.ChrtID, &It.TrackNumber, &It.Price, &It.Rid, &It.Name, &It.Sale, &It.Size, &It.TotalPrice, &It.NmID, &It.Brand, &It.Status)
		if err != nil {
			return err
		}
		if Cache[Or.OrderUID].OrderUID != Or.OrderUID {
			Or.Items = append(Or.Items, It)
			Cache[Or.OrderUID] = Or
		} else {
			exampOrder := Cache[Or.OrderUID]
			exampOrder.Items = append(exampOrder.Items, It)
			Cache[Or.OrderUID] = exampOrder
		}
	}
	return err
}

// GetByUID отвечает за получение Order по UID и возвращает Order и  был ли такой в кэшэ
func GetByUID(uid string) (Order, bool) {
	order, ok := Cache[uid]
	return order, ok
}

// JsonToObject переводит Json в struct Order
func JsonToObject(input string) (Order, error) {
	var order Order
	err := json.Unmarshal([]byte(input), &order)
	if order.Payment.Transaction == "" || order.OrderUID == "" ||
		order.TrackNumber == "" {
		return order, errors.New("некорректные данные")
	}
	return order, err
}

// ObjectToJson Переводит Order в string формата Json
func ObjectToJson(order Order) (string, error) {
	resp, err := json.MarshalIndent(order, "", "  ")
	return string(resp), err
}

func InsertData(db *pgx.Conn, order Order) error {
	err := insertPayment(db, order)
	if err != nil {
		return err
	}
	err = insertOrder(db, order)
	if err != nil {
		return err
	}

	err = insertItem(db, order)
	if err != nil {
		return err
	}

	id, err := insertDelivery(db, order)
	if err != nil {
		return err
	}
	err = insertOrderDelivery(db, order, id)
	if err != nil {
		return err
	}
	Cache[order.OrderUID] = order
	return nil
}

// Принимает соединение и order
func insertPayment(db *pgx.Conn, order Order) error {
	payment := order.Payment
	_, err := db.Exec(context.Background(), sqlInsertPayment,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDt,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee)
	return err
}

// Принимает соединение и order и добавляет в базу описание заказа
func insertOrder(db *pgx.Conn, order Order) error {
	_, err := db.Exec(context.Background(), sqlInsertOrder,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	return err
}

// Принимает соединение, trackNumber и order
func insertItem(db *pgx.Conn, order Order) error {
	items := order.Items
	for i := 0; i < len(items); i++ {
		_, err := db.Exec(context.Background(), sqlInsertItem,
			items[i].ChrtID,
			order.TrackNumber,
			items[i].Price,
			items[i].Rid,
			items[i].Name,
			items[i].Sale,
			items[i].Size,
			items[i].TotalPrice,
			items[i].NmID,
			items[i].Brand,
			items[i].Status)
		if err != nil {
			return err
		}
	}
	return nil
}

// Принимает соединение и order и возвращает deleveries.id
func insertDelivery(db *pgx.Conn, order Order) (int64, error) {
	var id int64
	err := db.QueryRow(context.Background(), sqlInsertDelivery,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email).Scan(&id)
	return id, err
}

// Принимает соединение и order_uid, deleveries.id
func insertOrderDelivery(db *pgx.Conn, order Order, deliveryID int64) error {
	_, err := db.Exec(context.Background(), sqlInsertOrderDelivery,
		order.OrderUID,
		deliveryID)
	return err
}
