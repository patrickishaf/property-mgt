## Input Data Models

These are models of the data the client will receive from the server

### Portfolio

```java
class Portfolio {
  private double totalRentExpected;
  private int activeTenantCount;
  private double totalRentCollected;
  private float percentageRentCollected;
  private double outstandingRentBalance;
  private int openInstallmentCount;
  private float totalOverdueAmount;
  private int overdueInstallmentCount;
}
```

### Due Payment

```java
class DuePayment {
  private String tenantFullName;
  private String propertyName;
  private String propertyUnitNumber;
  private float amount;
  private String dueDate;
}
```

### Payment

```java
class Payment {
  private String tenantFullName;
  private String propertyName;
  private String propertyUnitNumber;
  private float amountPaid;
  private String paymentDate;
  private String status: // 'paid' | 'partially_paid';
  private String reference;
  private String paymentMethod;
  private String status;
}
```

### Tenant

```java
class Tenant {
  private String firstName;
  private String lastName;
  private String email;
  private String phoneNumber;
  private int leaseId;
  private float expectedRent;
  private float outstandingRent;
}
```

### Lease

```java
class Lease {
  private DateTime startDate;
  private DateTime endDate;
  private float expectedRent;
  private float amountPaid;
  private float outstandingRent;
  private int propertyId;
  private int unitNumber;
}
```

### Installment

```java
class Installment {
  private float amountExpected;
  private float amountPaid;
  private float outstandingAmount;
  private Date dueDate;
  private String status;
}
```

### Receipt

```java
class PaymentReference {
  private String senderFullName;
  private String senderEmail;
  private String paymentDate;
  private String paymentMethod;
  private String amountPaid;
  private String intstallmentStatus;
  private String installmentAmount;
  private String installmentAmountPaid;
  private String previousPaymentAmount;
  private String outstandingInstallmentAmount;
  private String installmentDueDate;
}
```

### Property

```java
class Property {
  private String type;
  private String activityStatus;
  private String name;
  private int numberOfUnits;
  private int numberOfOccupiedUnits;
  private int totalOutstandingPayment;
  private float rentExpected;
  private float rentCollected;
  private String location;
  private String city;
  private String state;
  private String country;
  private int managerId;
  private float annualRentExpected;
}
```

### PropertyManager

```java
class PropertyManager {
  private String firstName;
  private String lastName;
  private String email;
  private String phoneNumber;
}
```

### Message

```java
class Message {
  private Date deliveredAt;
  private String channel;
  private String subject;
  private String body;
  private String senderName;
  private String numberOfRecipients;
}
```

### MessageRecipient

```java
class MessageRecipient {
  private String fullName;
  private String propertyUnitNumber;
}
```

### Complaint

```java
class Complaint {
  private String content;
  private Date createdAt;
  private String fromTenantName;
  private String fromTenantId;
  private String fromTenantUnitId;
  private String category;
  private String urgency;
  private String status;
}
```

### ComplaintUpdate

```java
class ComplaintUpdate {
  private Date createdAt;
  private String status;
  private String message;
  private String creatorFullName;
  private String creatorId;
}
```

### Issue

```java
class Issue {}
```