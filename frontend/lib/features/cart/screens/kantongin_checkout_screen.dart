import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:qr_flutter/qr_flutter.dart';

import '../../../core/payment/deep_link_launcher.dart';
import '../../../core/utils/currency_formatter.dart';
import '../../../shared/models/models.dart';

class KantonginCheckoutScreen extends StatefulWidget {
  final PaymentIntentModel paymentIntent;
  final int? orderId;

  KantonginCheckoutScreen({super.key, required CheckoutResult result})
      : paymentIntent = result.paymentIntent,
        orderId = result.order.id;

  const KantonginCheckoutScreen.fromIntent({
    super.key,
    required this.paymentIntent,
    this.orderId,
  });

  @override
  State<KantonginCheckoutScreen> createState() => _KantonginCheckoutScreenState();
}

class _KantonginCheckoutScreenState extends State<KantonginCheckoutScreen> {
  bool _opening = false;

  String get _paymentPayload {
    final link = widget.paymentIntent.deepLink.trim();
    if (link.isNotEmpty) return link;
    return 'kantongin://pay?token=${widget.paymentIntent.token}';
  }

  Future<void> _openKantongin() async {
    setState(() => _opening = true);
    final opened = await DeepLinkLauncher.openWallet(_paymentPayload);
    if (!mounted) return;
    setState(() => _opening = false);

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          opened
              ? 'Kantongin dibuka. Lanjutkan pembayaran dari aplikasi Kantongin.'
              : 'Kantongin belum bisa dibuka otomatis. Scan QR ini dari menu Scan QR Kantongin.',
        ),
        backgroundColor: opened ? Colors.green : Colors.orange,
      ),
    );
  }

  Future<void> _copyLink() async {
    await Clipboard.setData(ClipboardData(text: _paymentPayload));
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Link pembayaran disalin')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final intent = widget.paymentIntent;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Pembayaran Kantongin'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black,
        elevation: 0,
      ),
      backgroundColor: const Color(0xFFF5F5F5),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(20, 16, 20, 28),
        children: [
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Order berhasil dibuat', style: TextStyle(fontWeight: FontWeight.w800, fontSize: 18)),
                const SizedBox(height: 8),
                Text('Selesaikan pembayaran lewat Kantongin. Kalau akun Kantongin belum punya PIN transaksi, Kantongin akan meminta kamu membuat PIN terlebih dahulu.'),
                const SizedBox(height: 16),
                _row('Merchant', intent.merchant),
                if (widget.orderId != null) _row('Order', '#${widget.orderId}'),
                _row('Total', formatRupiah(intent.amount)),
                _row('Status', intent.status),
              ],
            ),
          ),
          const SizedBox(height: 18),
          Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
            ),
            child: Column(
              children: [
                const Text('QR pembayaran', style: TextStyle(fontWeight: FontWeight.w800, fontSize: 18)),
                const SizedBox(height: 12),
                QrImageView(data: _paymentPayload, size: 230, backgroundColor: Colors.white),
                const SizedBox(height: 12),
                Text('Scan QR ini dari menu Scan QR Kantongin.', style: TextStyle(color: Colors.grey[600]), textAlign: TextAlign.center),
              ],
            ),
          ),
          const SizedBox(height: 18),
          SizedBox(
            height: 50,
            child: ElevatedButton.icon(
              onPressed: _opening ? null : _openKantongin,
              style: ElevatedButton.styleFrom(backgroundColor: Colors.black, foregroundColor: Colors.white),
              icon: _opening
                  ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Icon(Icons.account_balance_wallet_rounded),
              label: Text(_opening ? 'Membuka Kantongin...' : 'Buka Kantongin'),
            ),
          ),
          const SizedBox(height: 10),
          OutlinedButton.icon(
            onPressed: _copyLink,
            icon: const Icon(Icons.copy_rounded),
            label: const Text('Salin link pembayaran'),
          ),
          const SizedBox(height: 10),
          TextButton(
            onPressed: () => Navigator.pushNamedAndRemoveUntil(context, '/orders', (route) => route.isFirst),
            child: const Text('Lihat daftar order'),
          ),
        ],
      ),
    );
  }

  Widget _row(String label, String value) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Row(
          children: [
            Expanded(child: Text(label, style: TextStyle(color: Colors.grey[600]))),
            Expanded(child: Text(value, textAlign: TextAlign.right, style: const TextStyle(fontWeight: FontWeight.w800))),
          ],
        ),
      );
}
