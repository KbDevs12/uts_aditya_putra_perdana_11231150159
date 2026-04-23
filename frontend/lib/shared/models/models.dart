class ProductModel {
  final int id;
  final String name;
  final String brand;
  final String description;
  final double price;
  final int stock;
  final String imageUrl;
  final String type;
  final int longevityHours;

  ProductModel({
    required this.id,
    required this.name,
    required this.brand,
    required this.description,
    required this.price,
    required this.stock,
    required this.imageUrl,
    required this.type,
    required this.longevityHours,
  });

  factory ProductModel.fromJson(Map<String, dynamic> json) => ProductModel(
    id: json['id'],
    name: json['name'] ?? '',
    brand: json['brand'] ?? '',
    description: json['description'] ?? '',
    price: (json['price'] as num).toDouble(),
    stock: json['stock'] ?? 0,
    imageUrl: json['image_url'] ?? '',
    type: json['type'] ?? '',
    longevityHours: json['longevity_hours'] ?? 0,
  );
}

class CartItemModel {
  final int id;
  final int cartId;
  final int productId;
  final int quantity;
  final double price;
  final String productName;
  final String productBrand;
  final String productImageUrl;

  CartItemModel({
    required this.id,
    required this.cartId,
    required this.productId,
    required this.quantity,
    required this.price,
    this.productName = '',
    this.productBrand = '',
    this.productImageUrl = '',
  });

  factory CartItemModel.fromJson(Map<String, dynamic> json) => CartItemModel(
    id: json['id'],
    cartId: json['cart_id'],
    productId: json['product_id'],
    quantity: json['quantity'],
    price: (json['price'] as num).toDouble(),
  );

  CartItemModel copyWith({
    String? productName,
    String? productBrand,
    String? productImageUrl,
  }) => CartItemModel(
    id: id,
    cartId: cartId,
    productId: productId,
    quantity: quantity,
    price: price,
    productName: productName ?? this.productName,
    productBrand: productBrand ?? this.productBrand,
    productImageUrl: productImageUrl ?? this.productImageUrl,
  );
}

class OrderModel {
  final int id;
  final int userId;
  final double totalPrice;
  final String status;

  OrderModel({
    required this.id,
    required this.userId,
    required this.totalPrice,
    required this.status,
  });

  factory OrderModel.fromJson(Map<String, dynamic> json) => OrderModel(
    id: json['id'],
    userId: json['user_id'],
    totalPrice: (json['total_price'] as num).toDouble(),
    status: json['status'] ?? 'pending',
  );
}

class OrderItemModel {
  final int id;
  final int orderId;
  final int productId;
  final int quantity;
  final double price;
  final String productName;
  final String productBrand;

  OrderItemModel({
    required this.id,
    required this.orderId,
    required this.productId,
    required this.quantity,
    required this.price,
    this.productName = '',
    this.productBrand = '',
  });

  factory OrderItemModel.fromJson(Map<String, dynamic> json) => OrderItemModel(
    id: json['id'],
    orderId: json['order_id'],
    productId: json['product_id'],
    quantity: json['quantity'],
    price: (json['price'] as num).toDouble(),
  );

  OrderItemModel copyWith({String? productName, String? productBrand}) =>
      OrderItemModel(
        id: id,
        orderId: orderId,
        productId: productId,
        quantity: quantity,
        price: price,
        productName: productName ?? this.productName,
        productBrand: productBrand ?? this.productBrand,
      );
}
