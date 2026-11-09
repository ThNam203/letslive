import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_models.dart';

/// Transactions per page; the backend caps `limit` at 20
/// (`backend/finance/handlers/utils/query.go`, web uses the same).
const int transactionsPageSize = 20;

/// The read-only wallet endpoints, mirroring `web/lib/api/wallet.ts`.
/// Deposits and payments are deliberately absent: they are out of scope for
/// the mobile app. Failures are values ([ApiResponse.success] false).
class WalletRepository {
  const WalletRepository(this._client);

  final ApiClient _client;

  /// `GET /wallet`: the signed-in user's balances.
  Future<ApiResponse<WalletOverview>> overview() =>
      _client.get(ApiEndpoints.wallet, parse: WalletOverview.parse);

  /// `GET /transactions`: [transactionsPageSize] per [page] (0-based),
  /// newest first.
  Future<ApiResponse<List<WalletTransaction>>> transactions(int page) =>
      _client.get(
        ApiEndpoints.transactions,
        query: {'page': page, 'limit': transactionsPageSize},
        parse: WalletTransaction.parseList,
      );

  /// `GET /currencies`.
  Future<ApiResponse<List<Currency>>> currencies() =>
      _client.get(ApiEndpoints.currencies, parse: Currency.parseList);
}
