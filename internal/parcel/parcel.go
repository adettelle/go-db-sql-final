package parcel

import (
	"database/sql"
	"log"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// добавление строки в таблицу parcel, используя данные из переменной p
func (s ParcelStore) Add(p Parcel) (int, error) {
	p.Status = ParcelStatusRegistered

	parcel, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	id, err := parcel.LastInsertId()
	if err != nil {
		return 0, err
	}

	// возвращаем идентификатор последней добавленной записи
	return int(id), nil
}

// чтение строки по заданному number
// Get: из таблицы должна вернуться только одна строка
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow("SELECT number, client, status, address, created_at from parcel WHERE number = :number",
		sql.Named("number", number))

	// заполняем объект Parcel данными из таблицы
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

// чтение строк из таблицы parcel по заданному client
// здесь из таблицы может вернуться несколько строк
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// заполняем срез Parcel данными из таблицы
	var res []Parcel
	for rows.Next() {
		parcel := Parcel{}
		err = rows.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, parcel)
	}

	return res, nil
}

// обновление статуса в таблице parcel
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("number", number),
		sql.Named("status", status))
	if err != nil {
		return err
	}

	return nil
}

// обновление адреса в таблице parcel
// менять адрес можно только если значение статуса registered
func (s ParcelStore) SetAddress(number int, address string) error {
	parcel, err := s.Get(number)
	if err != nil {
		log.Println(err)
	}
	if parcel.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
			sql.Named("number", number),
			sql.Named("address", address))
		if err != nil {
			return err
		}
	} else {
		log.Println(err)
	}
	return nil
}

// удаление строки из таблицы parcel
// удалять строку можно только если значение статуса registered
func (s ParcelStore) Delete(number int) error {
	p := Parcel{}
	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))

	err := row.Scan(&p.Status)
	if err != nil {
		return err
	}
	if p.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
		if err != nil {
			return err
		}
	} else {
		log.Println(err)
	}

	return nil
}
