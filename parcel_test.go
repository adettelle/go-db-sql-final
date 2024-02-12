package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/Yandex-Practicum/go-db-sql-final/parcel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() parcel.Parcel {
	return parcel.Parcel{
		Client:    1000,
		Status:    parcel.ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настраиваем подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := parcel.NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)

	// get
	// получаем только что добавленную посылку
	gotParcel, err := store.Get(parcelID)
	require.NoError(t, err)
	gotParcel.Number = parcel.Number
	assert.Equal(t, parcel, gotParcel)

	// delete
	// удаляем добавленную посылку
	// проверяем, что посылку больше нельзя получить из БД
	err = store.Delete(parcelID)
	require.NoError(t, err)

	got, err := store.Get(parcelID)
	require.Equal(t, sql.ErrNoRows, err)
	require.Empty(t, got)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := parcel.NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)

	// set address
	// обновляем адрес
	newAddress := "new test address"
	err = store.SetAddress(parcelID, newAddress)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и проверяем, что адрес обновился
	got, err := store.Get(parcelID)
	require.NoError(t, err)
	assert.Equal(t, newAddress, got.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := parcel.NewParcelStore(db)
	p := getTestParcel()

	// add
	// добавляем новую посылку в БД
	parcelID, err := store.Add(p)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)

	// set status
	// обновляем статус
	err = store.SetStatus(parcelID, parcel.ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и проверяем, что статус обновился
	got, err := store.Get(parcelID)
	require.NoError(t, err)
	assert.Equal(t, parcel.ParcelStatusSent, got.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := parcel.NewParcelStore(db)

	parcels := []parcel.Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]parcel.Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавдляем новую посылку в БД
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получаем список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)

	// проверяем, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, p := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// проверяем, что все посылки из storedParcels есть в parcelMap
		require.Equal(t, p, parcelMap[p.Number])
		// проверяем, что значения полей полученных посылок заполнены верно
		require.Equal(t, storedParcels, parcels)
	}
}
