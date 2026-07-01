import 'package:url_launcher/url_launcher.dart';

class DeepLinkLauncher {
  static Future<bool> openWallet(String deepLink) async {
    if (deepLink.trim().isEmpty) return false;
    final uri = Uri.tryParse(deepLink.trim());
    if (uri == null) return false;

    try {
      return await launchUrl(uri, mode: LaunchMode.externalApplication);
    } catch (_) {
      return false;
    }
  }
}
