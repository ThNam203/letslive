// ---------------------------------------------------------------------------
// Currency
// ---------------------------------------------------------------------------

enum CurrencyCode { spark, flare }

extension CurrencyCodeX on CurrencyCode {
  String get value => switch (this) {
    CurrencyCode.spark => 'SPARK',
    CurrencyCode.flare => 'FLARE',
  };

  static CurrencyCode fromString(String s) => switch (s.toUpperCase()) {
    'SPARK' => CurrencyCode.spark,
    'FLARE' => CurrencyCode.flare,
    _ => CurrencyCode.spark,
  };
}

class Currency {
  final String code;
  final String name;
  final int precision;

  const Currency({
    required this.code,
    required this.name,
    required this.precision,
  });

  factory Currency.fromJson(Map<String, dynamic> json) {
    return Currency(
      code: json['code'] as String,
      name: json['name'] as String,
      precision: json['precision'] as int,
    );
  }
}

// ---------------------------------------------------------------------------
// Account & Balance
// ---------------------------------------------------------------------------

enum AccountStatus { active, frozen, closed }

class Account {
  final String id;
  final String ownerId;
  final String type;
  final AccountStatus status;
  final String createdAt;
  final String updatedAt;

  const Account({
    required this.id,
    required this.ownerId,
    required this.type,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Account.fromJson(Map<String, dynamic> json) {
    return Account(
      id: json['id'] as String,
      ownerId: json['ownerId'] as String,
      type: json['type'] as String,
      status: AccountStatus.values.firstWhere(
        (e) => e.name == json['status'],
        orElse: () => AccountStatus.active,
      ),
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
    );
  }
}

class AccountBalance {
  final String accountId;
  final String currencyCode;
  final String balance;
  final String? lastEntryId;

  const AccountBalance({
    required this.accountId,
    required this.currencyCode,
    required this.balance,
    this.lastEntryId,
  });

  factory AccountBalance.fromJson(Map<String, dynamic> json) {
    return AccountBalance(
      accountId: json['accountId'] as String,
      currencyCode: json['currencyCode'] as String,
      balance: json['balance'] as String,
      lastEntryId: json['lastEntryId'] as String?,
    );
  }
}

class WalletOverview {
  final Account account;
  final List<AccountBalance> balances;

  const WalletOverview({required this.account, required this.balances});

  factory WalletOverview.fromJson(Map<String, dynamic> json) {
    return WalletOverview(
      account: Account.fromJson(json['account'] as Map<String, dynamic>),
      balances: (json['balances'] as List<dynamic>)
          .map((e) => AccountBalance.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  String balanceFor(CurrencyCode code) {
    final found = balances
        .where((b) => b.currencyCode == code.value)
        .firstOrNull;
    return found?.balance ?? '0';
  }
}

// ---------------------------------------------------------------------------
// Transaction
// ---------------------------------------------------------------------------

enum TransactionType { reward, purchase, trade, donate, refund, fee, adjustment }

enum TransactionStatus { created, processing, completed, failed, cancelled }

class LedgerEntry {
  final String id;
  final String transactionId;
  final String accountId;
  final String currencyCode;
  final String amount;
  final String createdAt;

  const LedgerEntry({
    required this.id,
    required this.transactionId,
    required this.accountId,
    required this.currencyCode,
    required this.amount,
    required this.createdAt,
  });

  factory LedgerEntry.fromJson(Map<String, dynamic> json) {
    return LedgerEntry(
      id: json['id'] as String,
      transactionId: json['transactionId'] as String,
      accountId: json['accountId'] as String,
      currencyCode: json['currencyCode'] as String,
      amount: json['amount'] as String,
      createdAt: json['createdAt'] as String,
    );
  }
}

class Transaction {
  final String id;
  final TransactionType type;
  final TransactionStatus status;
  final String? reference;
  final String? description;
  final String actorId;
  final String createdAt;
  final String updatedAt;
  final List<LedgerEntry>? entries;

  const Transaction({
    required this.id,
    required this.type,
    required this.status,
    this.reference,
    this.description,
    required this.actorId,
    required this.createdAt,
    required this.updatedAt,
    this.entries,
  });

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'] as String,
      type: TransactionType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => TransactionType.purchase,
      ),
      status: TransactionStatus.values.firstWhere(
        (e) => e.name == json['status'],
        orElse: () => TransactionStatus.created,
      ),
      reference: json['reference'] as String?,
      description: json['description'] as String?,
      actorId: json['actorId'] as String,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
      entries: (json['entries'] as List<dynamic>?)
          ?.map((e) => LedgerEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  double get netAmount {
    if (entries == null || entries!.isEmpty) return 0;
    return entries!.fold(0.0, (sum, e) => sum + double.parse(e.amount));
  }
}

// ---------------------------------------------------------------------------
// Payment
// ---------------------------------------------------------------------------

enum PaymentStatus { pending, processing, completed, failed, cancelled }

enum PaymentProvider { stripe, paypal }

class Payment {
  final String id;
  final String? transactionId;
  final PaymentProvider provider;
  final String? providerReference;
  final String currencyCode;
  final String amount;
  final PaymentStatus status;
  final String createdAt;
  final String updatedAt;

  const Payment({
    required this.id,
    this.transactionId,
    required this.provider,
    this.providerReference,
    required this.currencyCode,
    required this.amount,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Payment.fromJson(Map<String, dynamic> json) {
    return Payment(
      id: json['id'] as String,
      transactionId: json['transactionId'] as String?,
      provider: PaymentProvider.values.firstWhere(
        (e) => e.name == json['provider'],
        orElse: () => PaymentProvider.stripe,
      ),
      providerReference: json['providerReference'] as String?,
      currencyCode: json['currencyCode'] as String,
      amount: json['amount'] as String,
      status: PaymentStatus.values.firstWhere(
        (e) => e.name == json['status'],
        orElse: () => PaymentStatus.pending,
      ),
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
    );
  }
}

class DepositResponse {
  final Payment payment;
  final String checkoutUrl;

  const DepositResponse({required this.payment, required this.checkoutUrl});

  factory DepositResponse.fromJson(Map<String, dynamic> json) {
    return DepositResponse(
      payment: Payment.fromJson(json['payment'] as Map<String, dynamic>),
      checkoutUrl: json['checkoutUrl'] as String,
    );
  }
}
