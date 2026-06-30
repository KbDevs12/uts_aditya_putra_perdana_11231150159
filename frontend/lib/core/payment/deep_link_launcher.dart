import 'package:url_launcher/url_launcher.dart';

class DeepLinkLauncher {
  static Future<bool> openWallet(String deepLink) async {
    final uri = Uri.parse(deepLink);
    if (await canLaunchUrl(uri)) {
      return launchUrl(uri, mode: LaunchMode.externalApplication);
    }
    return false;
  }
}
