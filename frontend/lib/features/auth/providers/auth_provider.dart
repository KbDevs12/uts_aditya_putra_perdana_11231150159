import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:dio/dio.dart';
import 'package:frontend/core/api/api_client.dart';
import 'package:frontend/core/constant/app_constant.dart';

enum AuthStatus { initial, loading, authenticated, unauthenticated, error }

class AuthProvider extends ChangeNotifier {
  final _api = ApiClient();
  final _firebaseAuth = FirebaseAuth.instance;

  AuthStatus _status = AuthStatus.initial;
  String? _errorMessage;

  AuthStatus get status => _status;
  String? get errorMessage => _errorMessage;

  Future<void> checkAuth() async {
    final token = await _api.getToken();
    _status = token != null
        ? AuthStatus.authenticated
        : AuthStatus.unauthenticated;
    notifyListeners();
  }

  Future<bool> register(String email, String password, String name) async {
    _status = AuthStatus.loading;
    _errorMessage = null;
    notifyListeners();

    try {
      final cred = await _firebaseAuth.createUserWithEmailAndPassword(
        email: email,
        password: password,
      );

      await cred.user!.sendEmailVerification();

      final idToken = await cred.user!.getIdToken(true);

      await _api.dio.post(
        AppConstant.register,
        data: {'token': idToken, 'name': name},
      );

      await _firebaseAuth.signOut();

      _status = AuthStatus.unauthenticated;
      notifyListeners();
      return true;
    } on FirebaseAuthException catch (e) {
      _errorMessage = _firebaseErrorMessage(e.code);
      _status = AuthStatus.error;
      notifyListeners();
      return false;
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Registration failed';
      _status = AuthStatus.error;
      notifyListeners();
      return false;
    }
  }

  Future<bool> login(String email, String password) async {
    _status = AuthStatus.loading;
    _errorMessage = null;
    notifyListeners();

    try {
      final cred = await _firebaseAuth.signInWithEmailAndPassword(
        email: email,
        password: password,
      );

      if (!cred.user!.emailVerified) {
        await _firebaseAuth.signOut();
        _errorMessage = 'Please verify your email first';
        _status = AuthStatus.error;
        notifyListeners();
        return false;
      }

      final idToken = await cred.user!.getIdToken(true);

      final response = await _api.dio.post(
        AppConstant.login,
        data: {'token': idToken},
      );

      final jwt = response.data['access_token'];
      await _api.saveToken(jwt);

      _status = AuthStatus.authenticated;
      notifyListeners();
      return true;
    } on FirebaseAuthException catch (e) {
      _errorMessage = _firebaseErrorMessage(e.code);
      _status = AuthStatus.error;
      notifyListeners();
      return false;
    } on DioException catch (e) {
      _errorMessage = e.response?.data['error'] ?? 'Login failed';
      _status = AuthStatus.error;
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    await _firebaseAuth.signOut();
    await _api.clearToken();
    _status = AuthStatus.unauthenticated;
    notifyListeners();
  }

  String _firebaseErrorMessage(String code) {
    switch (code) {
      case 'email-already-in-use':
        return 'Email already in use';
      case 'invalid-email':
        return 'Invalid email address';
      case 'weak-password':
        return 'Password is too weak (min 6 characters)';
      case 'user-not-found':
      case 'wrong-password':
      case 'invalid-credential':
        return 'Invalid email or password';
      case 'too-many-requests':
        return 'Too many attempts, please try again later';
      default:
        return 'Authentication error: $code';
    }
  }
}
