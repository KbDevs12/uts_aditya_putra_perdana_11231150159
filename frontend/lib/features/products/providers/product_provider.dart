import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import 'package:frontend/core/api/api_client.dart';
import 'package:frontend/core/constant/app_constant.dart';
import 'package:frontend/shared/models/models.dart';

class ProductProvider extends ChangeNotifier {
  final _api = ApiClient();

  List<ProductModel> _products = [];
  ProductModel? _selectedProduct;
  bool _isLoading = false;
  String? _errorMessage;

  List<ProductModel> get products => _products;
  ProductModel? get selectedProduct => _selectedProduct;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;

  Future<void> fetchProducts() async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final response = await _api.dio.get(AppConstant.products);
      _products = (response.data as List)
          .map((e) => ProductModel.fromJson(e))
          .toList();
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to load products';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> fetchProductDetail(int id) async {
    _isLoading = true;
    _selectedProduct = null;
    _errorMessage = null;
    notifyListeners();

    try {
      final response = await _api.dio.get('${AppConstant.products}/$id');
      _selectedProduct = ProductModel.fromJson(response.data);
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Failed to load product';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }
}
