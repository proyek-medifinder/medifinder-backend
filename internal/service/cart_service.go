package service

import (
	"errors"

	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type CartService struct {
	Gateway  PaymentGateway
	CartRepo *repository.CartRepository
	ObatRepo *repository.ObatRepository
}

func (s *CartService) GetCart(userID string) (interface{}, error) {

	cart, err := s.CartRepo.FindByUser(userID)
	if err != nil {
		return nil, errors.New("cart kosong")
	}

	items, err := s.CartRepo.GetCartItems(cart.ID.String())
	if err != nil {
		return nil, err
	}

	var total float64 = 0

	var result []map[string]interface{}

	for _, item := range items {

		subtotal := item.Harga * float64(item.Jumlah)
		total += subtotal

		result = append(result, map[string]interface{}{
			"item_id":  item.ID,
			"obat_id":  item.ObatID,
			"nama":     item.Nama,
			"harga":    item.Harga,
			"jumlah":   item.Jumlah,
			"subtotal": subtotal,
		})

	}

	return map[string]interface{}{
		"items": result,
		"total": total,
	}, nil
}
