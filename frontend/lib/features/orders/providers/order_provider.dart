import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import 'package:frontend/core/api/api_client.dart';
import 'package:frontend/core/constant/app_constant.dart';
import 'package:frontend/shared/models/models.dart';

class OrderProvider extends ChangeNotifier {
  final _api = ApiClient();

  List<OrderModel> _orders = [];
  OrderModel? _selectedOrder;
  List<OrderItemModel> _selectedOrderItems = [];
  bool _isLoading = false;
  String? _errorMessage;

  List<OrderModel> get orders => _orders;
  OrderModel? get selectedOrder => _selectedOrder;
  List<OrderItemModel> get selectedOrderItems => _selectedOrderItems;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;

  Future<bool> checkout() async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      await _api.dio.post(AppConstant.checkout);
      await fetchOrders();
      return true;
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Checkout failed';
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<void> fetchOrders() async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final response = await _api.dio.get(AppConstant.orders);
      _orders = (response.data as List)
          .map((e) => OrderModel.fromJson(e))
          .toList();
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to load orders';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> fetchOrderDetail(int id) async {
    _isLoading = true;
    _selectedOrder = null;
    _selectedOrderItems = [];
    _errorMessage = null;
    notifyListeners();

    try {
      final response = await _api.dio.get('${AppConstant.orders}/$id');
      _selectedOrder = OrderModel.fromJson(response.data['order']);
      final rawItems = (response.data['items'] as List)
          .map((e) => OrderItemModel.fromJson(e))
          .toList();

      // Enrich dengan nama produk
      _selectedOrderItems = await _enrichWithProductData(rawItems);
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to load order';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<List<OrderItemModel>> _enrichWithProductData(
    List<OrderItemModel> items,
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
        );
      } catch (_) {
        return item;
      }
    });
    return Future.wait(futures);
  }
}
