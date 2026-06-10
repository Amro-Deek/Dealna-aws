# Dealna Rating System Specifications & Mobile Guide

This document contains the high-level functional specifications, mathematical models, performance optimizations, and integration guidelines for the Dealna Rating System. It is intended for use in the graduation project report and as a strict guide for the Flutter Mobile Developers.

---

## 1. Core Rules & Constraints
1. **Scope:** Only completed **Purchases** are eligible for ratings. Giveaways (Queue system) are excluded from the rating system.
2. **Subject:** Ratings are explicitly for the **User** (the seller or buyer), not the item itself.
3. **The 14-Day Rule (Grace Period vs Mandatory):**
   - After a purchase is marked `COMPLETED`, the buyer has a **14-day grace period** to optionally submit a rating.
   - On exactly Day 14, the rating becomes **Mandatory**. The rating opportunity *never expires*. The user is restricted from using the platform until they clear their pending mandatory ratings.
4. **Reminders:** The backend automatically dispatches Firebase Push Notifications on **Day 3** and **Day 10** post-completion to remind the buyer to rate the transaction.

## 2. Mathematical Model (Bayesian Average)
To protect users from malicious reviews and ensure system stability, Dealna employs a Bayesian Average for calculating user ratings, heavily weighted to prevent manipulation.

**Formula:**
$$ R = \frac{n \times \bar{r} + m \times C}{n + m} $$

- **$n$**: Total number of ratings the specific user has received.
- **$\bar{r}$**: The simple arithmetic mean of the user's ratings.
- **$C$ (Global Average)**: Dynamically calculated based on the average of all ratings ever submitted to the platform.
- **$m$ (Smoothing Constant)**: Set strictly to **10**.

**Cold Start Problem:**
If the platform has fewer than 100 total ratings across all users, $C$ is hardcoded to **4.0** to prevent early chaotic fluctuations. Once the platform surpasses 100 ratings, $C$ becomes fully dynamic.

**Why $m=10$?**
A smoothing constant of 10 means a user's rating is a 50/50 split between their actual performance and the platform's global average until they complete exactly 10 transactions. This protects new users from being ruined by a single 1-star review, while making it mathematically impossible for a user to reach a 4.5+ rating using fake reviews from a few friends.

## 3. Dynamic Account Privileges
A user's Bayesian rating directly and instantly affects their daily listing limit:
- **Rating $\ge 4.5$**: Maximum privileges (Up to **10** listings per day).
- **$3.0 \le$ Rating $< 4.5$**: Standard privileges (Up to **7** listings per day).
- **Rating $< 3.0$**: Restricted privileges (Up to **4** listings per day).

## 4. Performance Optimizations (Backend)
To ensure the rating system introduces exactly **zero latency** and **zero blocking** to the system:
1. **Database Caching:** Bayesian averages are not calculated on-the-fly. The backend stores three cache columns directly on the `User` table: `total_ratings`, `sum_ratings`, and `bayesian_rating`. When a user loads the app, the system instantly reads the pre-calculated number.
2. **Asynchronous Processing:** All reminder checks and Push Notifications run in isolated asynchronous Go routines.
3. **Global Average Caching:** Calculating the global average ($C$) requires a massive SQL `SUM()/COUNT()` over thousands of rows. Instead of running this on every rating, it is cached in a specialized `sys_config` table and updated asynchronously via a background worker.

---

## 5. Mobile Developer Integration Guide

Mobile developers must integrate two new endpoints to enforce the rating system rules.

### 5.1 Submitting a Rating
When the buyer wants to rate a transaction, call this endpoint.

**Endpoint:** `POST /api/v1/transactions/{transactionId}/rate`
**Headers:** `Authorization: Bearer <token>`
**Body:**
```json
{
  "stars": 5,
  "comment": "Great seller, arrived on time!"
}
```

### 5.2 Mandatory Ratings (The 14-Day Lockout)
Every time the user opens the mobile application (or navigates to the Home screen), the Flutter app **MUST** call the pending ratings API.

**Endpoint:** `GET /api/v1/users/me/pending-ratings`
**Headers:** `Authorization: Bearer <token>`
**Response:**
```json
[
  {
    "transaction_id": "8fa21...",
    "item_id": "4bc45...",
    "item_title": "MacBook Air M1",
    "seller_id": "2bc99...",
    "seller_name": "Amro Al-Deek",
    "days_since_completion": 14
  }
]
```

**Mobile UX Flow for Mandatory Ratings:**
1. If this API returns an empty array `[]`, allow the user to proceed normally.
2. If this API returns **any objects**, the mobile app **MUST** immediately display a full-screen pop-up or modal.
3. **UI Text:** Use the data from the API to display a highly specific message:
   > *"Please rate your recent purchase of **[item_title]** from **[seller_name]**."*
4. **Enforcement:** The mobile app must **hide the close button** or disable background tapping for this modal. The user cannot dismiss it. The only way forward is to submit a rating via the `POST /rate` endpoint, which will clear the lock.

### 5.3 Fetching a User's Public Reviews
When a user taps on another user's profile, you can fetch a list of all the public reviews left for that specific seller.

**Endpoint:** `GET /api/v1/users/{userId}/ratings`
**Authentication:** No authentication required (Public endpoint).
**Query Parameters (Optional):**
- `limit`: Number of reviews to fetch (default: 20)
- `offset`: Pagination offset (default: 0)

**Response:**
```json
[
  {
    "rating_id": "8fa21...",
    "stars": 5,
    "comment": "Great seller, met me at the library on time!",
    "rater_name": "Amro Al-Deek",
    "created_at": "2026-06-10T12:00:00Z"
  }
]
```

### 5.4 Automated Rating Reminders (Push Notifications)
The backend runs a daily background worker. Exactly **3 days** after a transaction is marked as COMPLETED, if the buyer has not left a rating, the backend sends an FCM push notification to the buyer.

**FCM Payload (Data):**
```json
{
  "notification_type": "RATING_REMINDER",
  "item_id": "<the_item_id>",
  "item_title": "<automatically_fetched>",
  "unread_count": "<total_unread_notifications>"
}
```
**Mobile Action:**
When the user taps this notification, read the `item_id` and navigate them to the screen where they can leave a rating for that transaction.
