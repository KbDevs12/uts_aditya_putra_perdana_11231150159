import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import 'package:frontend/core/api/api_client.dart';
import 'package:frontend/core/constant/app_constant.dart';
import 'package:frontend/shared/models/models.dart';

class CartProvider extends ChangeNotifier {
  final _api = ApiClient();

  List<CartItemModel> _items = [];
  bool _isLoading = false;
  String? _errorMessage;

  List<CartItemModel> get items => _items;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;

  double get totalPrice =>
      _items.fold(0, (sum, item) => sum + (item.price * item.quantity));

  int get totalItems => _items.fold(0, (sum, item) => sum + item.quantity);

  Future<void> fetchCart() async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final response = await _api.dio.get(AppConstant.cart);
      final rawItems = (response.data as List)
          .map((e) => CartItemModel.fromJson(e))
          .toList();

      _items = await _enrichWithProductData(rawItems);
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to load cart';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<List<CartItemModel>> _enrichWithProductData(
    List<CartItemModel> items,
  ) async {
    final futures = items.map((item) async {
      try {
        final res = await _api.dio.get(
          '${AppConstant.products}/${item.productId}',
        );
        final product = ProductModel.fromJson(res.data);
        return item.copyWith(
          productName: product.name,
          productBrand: product.brand,
          productImageUrl: product.imageUrl,
        );
      } catch (_) {
        return item; // fallback tanpa nama
      }
    });
    return Future.wait(futures);
  }

  Future<bool> addToCart(int productId, int quantity) async {
    try {
      await _api.dio.post(
        AppConstant.cart,
        data: {'product_id': productId, 'quantity': quantity},
      );
      await fetchCart();
      return true;
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to add to cart';
      notifyListeners();
      return false;
    }
  }

  Future<bool> removeItem(int cartItemId) async {
    try {
      await _api.dio.delete('${AppConstant.cart}/$cartItemId');
      _items.removeWhere((item) => item.id == cartItemId);
      notifyListeners();
      return true;
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to remove item';
      notifyListeners();
      return false;
    }
  }

  Future<bool> clearCart() async {
    try {
      await _api.dio.delete(AppConstant.cart);
      _items = [];
      notifyListeners();
      return true;
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to clear cart';
      notifyListeners();
      return false;
    }
  }

  void reset() {
    _items = [];
    notifyListeners();
  }
}
