/// Response models of the finance service's wallet endpoints, typed after
/// `web/types/wallet.ts` (only the fields the app shows). Amounts are parsed
/// into [DecimalAmount], never `double`.
library;

import 'package:flutter/foundation.dart';
import 'package:letslive_mobile/core/network/json_fields.dart';
import 'package:letslive_mobile/features/wallet/domain/decimal_amount.dart';

/// One currency balance of the wallet (`balances[]` of `GET /wallet`).
@immutable
class WalletBalance {
  const WalletBalance({required this.currencyCode, required this.balance});

  factory WalletBalance.fromJson(Map<String, Object?> json) {
    final f = JsonFields('WalletBalance', json);
    return WalletBalance(
      currencyCode: f.string('currencyCode'),
      balance: _amount(f, 'balance'),
    );
  }

  /// `SPARK` or `FLARE`.
  final String currencyCode;

  /// What the user owns of it.
  final DecimalAmount balance;
}

/// `GET /wallet`: the balances of the signed-in user.
@immutable
class WalletOverview {
  const WalletOverview({required this.balances});

  factory WalletOverview.fromJson(Map<String, Object?> json) => WalletOverview(
    balances: parseJsonList(
      json['balances'],
      'WalletBalance',
      WalletBalance.fromJson,
    ),
  );

  /// `ParseData<WalletOverview>` for `ApiClient`.
  static WalletOverview parse(Object? data) =>
      parseJsonObject(data, 'WalletOverview', WalletOverview.fromJson);

  /// A balance per currency the user has an entry for.
  final List<WalletBalance> balances;

  /// The balance of [currencyCode]; zero when the wallet has no entry yet
  /// (web shows `0` the same way).
  DecimalAmount balanceOf(String currencyCode) {
    for (final balance in balances) {
      if (balance.currencyCode == currencyCode) return balance.balance;
    }
    return DecimalAmount.zero;
  }
}

/// One ledger movement of the user's own account in a transaction.
@immutable
class LedgerEntry {
  const LedgerEntry({required this.currencyCode, required this.amount});

  factory LedgerEntry.fromJson(Map<String, Object?> json) {
    final f = JsonFields('LedgerEntry', json);
    return LedgerEntry(
      currencyCode: f.string('currencyCode'),
      amount: _amount(f, 'amount'),
    );
  }

  /// Currency of the movement.
  final String currencyCode;

  /// Positive for a credit, negative for a debit.
  final DecimalAmount amount;
}

/// A wallet transaction (`GET /transactions`). The backend only returns the
/// entries of the signed-in user's own account.
@immutable
class WalletTransaction {
  const WalletTransaction({
    required this.id,
    required this.type,
    required this.status,
    required this.createdAt,
    required this.entries,
  });

  factory WalletTransaction.fromJson(Map<String, Object?> json) {
    final f = JsonFields('WalletTransaction', json);
    return WalletTransaction(
      id: f.string('id'),
      type: f.string('type'),
      status: f.string('status'),
      createdAt: f.date('createdAt'),
      // Go sends `null` for a transaction without entries.
      entries: json['entries'] == null
          ? const []
          : parseJsonList(json['entries'], 'LedgerEntry', LedgerEntry.fromJson),
    );
  }

  /// `ParseData<List<WalletTransaction>>` for `ApiClient`.
  static List<WalletTransaction> parseList(Object? data) =>
      parseJsonList(data, 'WalletTransaction', WalletTransaction.fromJson);

  /// The transaction's id.
  final String id;

  /// `reward`, `purchase`, `deposit`, `trade`, `donate`, `refund`, `fee` or
  /// `adjustment`.
  final String type;

  /// `created`, `processing`, `completed`, `failed` or `cancelled`.
  final String status;

  /// When it happened.
  final DateTime createdAt;

  /// The movements of the user's account.
  final List<LedgerEntry> entries;

  /// What the user's account gained (or lost, when negative): the exact sum
  /// of [entries], like web's `computeNetAmount` but without floating point.
  DecimalAmount get netAmount =>
      entries.fold(DecimalAmount.zero, (sum, e) => sum + e.amount);
}

/// A currency of the platform (`GET /currencies`).
@immutable
class Currency {
  const Currency({
    required this.code,
    required this.name,
    required this.precision,
  });

  factory Currency.fromJson(Map<String, Object?> json) {
    final f = JsonFields('Currency', json);
    return Currency(
      code: f.string('code'),
      name: f.string('name'),
      precision: f.integer('precision'),
    );
  }

  /// `ParseData<List<Currency>>` for `ApiClient`.
  static List<Currency> parseList(Object? data) =>
      parseJsonList(data, 'Currency', Currency.fromJson);

  /// `SPARK` or `FLARE`.
  final String code;

  /// English name.
  final String name;

  /// Fraction digits the server keeps for it.
  final int precision;
}

DecimalAmount _amount(JsonFields f, String key) {
  try {
    return DecimalAmount.parse(f.string(key));
  } on FormatException catch (e) {
    throw FormatException('${f.model}.$key: ${e.message}');
  }
}
