package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

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
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// Подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id

	fmt.Printf("Добавлена посылка № %d на адрес %s\n", parcel.Number, parcel.Address)

	// Получаем посылку
	retrieved, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, retrieved.Client)
	require.Equal(t, parcel.Status, retrieved.Status)
	require.Equal(t, parcel.Address, retrieved.Address)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
	fmt.Printf("Посылка № %d успешно удалена\n", id)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// Подключение к БД
	db, err := sql.Open("sqlite", "tracker.db") // Исправил "sqlite" -> "sqlite3"
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id

	fmt.Printf(" Добавлена посылка № %d на адрес %s\n", parcel.Number, parcel.Address)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	resReceived, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, resReceived.Address)

	fmt.Printf(" Адрес посылки № %d обновлён на %s\n", parcel.Number, newAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = id

	fmt.Printf(" Добавлена посылка № %d\n", parcel.Number)

	newStatus := ParcelStatusSent
	err = store.SetStatus(parcel.Number, newStatus)
	require.NoError(t, err)

	retrieved, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newStatus, retrieved.Status)

	fmt.Printf("✅ Статус посылки № %d обновлён на %s\n", parcel.Number, newStatus)
}

func TestGetByClient(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// Создаём тестовые посылки с одним client
	client := randRange.Intn(10_000_000)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	for i := range parcels {
		parcels[i].Client = client
	}

	// Добавляем посылки в БД
	parcelMap := make(map[int]Parcel)
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// Получаем посылки по client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// Проверяем, что каждая полученная посылка соответствует ожидаемой
	for _, storedParcel := range storedParcels {
		expectedParcel, exists := parcelMap[storedParcel.Number]
		require.True(t, exists)
		require.Equal(t, expectedParcel.Client, storedParcel.Client)
		require.Equal(t, expectedParcel.Status, storedParcel.Status)
		require.Equal(t, expectedParcel.Address, storedParcel.Address)
	}

	fmt.Printf("Успешно получены %d посылки для клиента %d\n", len(storedParcels), client)
}
