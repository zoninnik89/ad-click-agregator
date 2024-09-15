package tests

import (
	"context"
	"github.com/stretchr/testify/assert"
	mocks "github.com/zoninnik89/ad-click-aggregator/aggregator/tests/mocks"
	protoBuff "github.com/zoninnik89/commons/api"
	gomock "go.uber.org/mock/gomock"
	"testing"
)

func TestService_StoreClick(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Мокаем зависимости
	mockStore := mocks.NewMockStoreInterface(ctrl)
	//mockGateway := mocks.NewMockAdsGatewayInterface(ctrl)

	// Создаем сервис с моками
	service := mocks.NewMockAggregatorService(ctrl)

	// Пример запроса
	request := &protoBuff.SendClickRequest{AdID: "ad123"}

	// Мокаем добавление клика
	mockStore.EXPECT().AddClick("ad123")

	// Вызываем тестируемую функцию
	result, err := service.StoreClick(context.TODO(), request)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, "ad123", result.AdID)
	//assert.Equal(t, int64(1620000000), result.Timestamp)
	assert.True(t, result.IsAccepted)
}

//func TestService_GetClicksCounter(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockStore := mocks.NewMockStoreInterface(ctrl)
//	//mockGateway := mocks.NewMockAdsGatewayInterface(ctrl)
//
//	// Создаем сервис с моками
//	service := mocks.NewMockAggregatorService(ctrl)
//
//	// Создаем пример запроса
//	request := &protoBuff.GetClicksCounterRequest{AdId: "ad123"}
//
//	// Устанавливаем поведение мока
//	mockStore.EXPECT().GetCount("ad123").Return(10)
//
//	// Вызываем тестируемую функцию
//	result, err := service.GetClicksCounter(context.TODO(), request)
//
//	// Проверяем результат
//	assert.NoError(t, err)
//	assert.Equal(t, int32(10), result.Count)
//}
