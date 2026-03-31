import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/wallet.dart';

class WalletRepository {
  final ApiClient _client;

  WalletRepository(this._client);

  Future<ApiResponse<WalletOverview>> getMyWallet() {
    return _client.get(
      ApiEndpoints.financeWallet,
      fromJsonT: (json) =>
          WalletOverview.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<List<Currency>>> getCurrencies() {
    return _client.get(
      ApiEndpoints.financeCurrencies,
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Currency.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<Transaction>>> getTransactions({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.financeTransactions,
      queryParameters: {'page': page, 'page_size': pageSize},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Transaction.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<DepositResponse>> createDeposit({
    required String provider,
    required String currencyCode,
    required String amount,
  }) {
    return _client.post(
      ApiEndpoints.financeDeposits,
      data: {
        'provider': provider,
        'currencyCode': currencyCode,
        'amount': amount,
      },
      fromJsonT: (json) =>
          DepositResponse.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<List<Payment>>> getPayments({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.financePayments,
      queryParameters: {'page': page, 'page_size': pageSize},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Payment.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
