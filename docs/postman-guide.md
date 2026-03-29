# Postman Guide — E-commerce API

**Base URL:** `http://localhost:8080`

Все ответы имеют формат:
```json
{ "success": true, "data": { ... } }
{ "success": false, "error": "сообщение об ошибке" }
```

---

## Настройка переменных окружения в Postman

Создай Environment с переменными:

| Variable        | Initial Value                  |
|-----------------|-------------------------------|
| `base_url`      | `http://localhost:8080`       |
| `access_token`  | _(заполняется автоматически)_ |
| `refresh_token` | _(заполняется автоматически)_ |

В запросах используй `{{base_url}}`, `{{access_token}}` и т.д.

Чтобы токены сохранялись автоматически, добавь в **Tests** логина/регистрации:
```javascript
const data = pm.response.json().data;
pm.environment.set("access_token",  data.tokens.access_token);
pm.environment.set("refresh_token", data.tokens.refresh_token);
```

---

## 1. Health Check

| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/health` |
| Auth | Нет |

**Response 200:**
```json
{ "status": "ok", "service": "ecommerce", "db": "ok" }
```

---

## 2. Auth

### Регистрация
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/register` |
| Body | `raw → JSON` |

```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "first_name": "Иван",
  "last_name": "Иванов"
}
```

**Response 201:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "Иван",
      "last_name": "Иванов",
      "role": "customer",
      "created_at": "2026-03-30T10:00:00Z"
    },
    "tokens": {
      "access_token": "eyJ...",
      "refresh_token": "abc123...",
      "expires_in": 900
    }
  }
}
```

---

### Логин
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/login` |
| Body | `raw → JSON` |

```json
{
  "email": "admin@example.com",
  "password": "Admin1234!"
}
```

**Response 200** — тот же формат что и регистрация.

---

### Обновление токена
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/refresh` |
| Body | `raw → JSON` |

```json
{
  "refresh_token": "{{refresh_token}}"
}
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "xyz789...",
    "expires_in": 900
  }
}
```

---

### Logout
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/auth/logout` |
| Body | `raw → JSON` |

```json
{
  "refresh_token": "{{refresh_token}}"
}
```

**Response 204** — пустой body.

---

## 3. Авторизация в Postman

Для всех защищённых запросов:
- Вкладка **Authorization**
- Type: `Bearer Token`
- Token: `{{access_token}}`

---

## 4. Продукты

### Список продуктов (публичный)
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/products` |
| Auth | Нет |

**Query параметры (все опциональны):**

| Параметр      | Тип     | Пример          |
|---------------|---------|-----------------|
| `page`        | int     | `1`             |
| `limit`       | int     | `20`            |
| `search`      | string  | `iPhone`        |
| `category_id` | int     | `1`             |
| `min_price`   | float   | `100`           |
| `max_price`   | float   | `1000`          |

**Response 200:**
```json
{
  "success": true,
  "data": [ { "id": 1, "name": "iPhone 15 Pro", "price": 699.99, ... } ],
  "meta": { "page": 1, "limit": 20, "total": 14 }
}
```

---

### Получить продукт по ID (публичный)
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/products/1` |
| Auth | Нет |

---

### Создать продукт (admin / seller)
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/products` |
| Auth | Bearer Token |

```json
{
  "category_id": 1,
  "name": "Новый товар",
  "description": "Описание товара",
  "price": 199.99,
  "stock": 50,
  "image_url": ""
}
```

---

### Обновить продукт (admin / seller-владелец)
| | |
|---|---|
| Method | `PUT` |
| URL | `{{base_url}}/api/v1/products/1` |
| Auth | Bearer Token |

```json
{
  "name": "Обновлённое название",
  "description": "Новое описание",
  "price": 249.99,
  "stock": 30,
  "image_url": ""
}
```

---

### Удалить продукт (admin / seller-владелец)
| | |
|---|---|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/products/1` |
| Auth | Bearer Token |

**Response 204.**

---

## 5. Категории

### Список категорий (публичный)
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/categories` |

---

### Создать категорию (admin)
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/categories` |
| Auth | Bearer Token (admin) |

```json
{
  "name": "Игрушки",
  "slug": "toys"
}
```

---

### Обновить категорию (admin)
| | |
|---|---|
| Method | `PUT` |
| URL | `{{base_url}}/api/v1/categories/1` |
| Auth | Bearer Token (admin) |

```json
{
  "name": "Электроника",
  "slug": "electronics"
}
```

---

### Удалить категорию (admin)
| | |
|---|---|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/categories/1` |
| Auth | Bearer Token (admin) |

---

## 6. Корзина

### Посмотреть корзину
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/cart` |
| Auth | Bearer Token |

---

### Добавить товар в корзину
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/cart/items` |
| Auth | Bearer Token |

```json
{
  "product_id": 1,
  "quantity": 2
}
```

---

### Обновить количество товара
| | |
|---|---|
| Method | `PUT` |
| URL | `{{base_url}}/api/v1/cart/items/1` |
| Auth | Bearer Token |

```json
{
  "quantity": 3
}
```

---

### Удалить товар из корзины
| | |
|---|---|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/cart/items/1` |
| Auth | Bearer Token |

---

### Очистить корзину
| | |
|---|---|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/cart` |
| Auth | Bearer Token |

---

### Оформить заказ из корзины (checkout)
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/cart/checkout` |
| Auth | Bearer Token |
| Body | Пустой |

> Опционально: добавь заголовок `Idempotency-Key: <uuid>` для защиты от двойного списания.

---

## 7. Заказы

### Создать заказ напрямую
| | |
|---|---|
| Method | `POST` |
| URL | `{{base_url}}/api/v1/orders` |
| Auth | Bearer Token |
| Header | `Idempotency-Key: <uuid>` _(опционально)_ |

```json
{
  "items": [
    { "product_id": 1, "quantity": 1 },
    { "product_id": 5, "quantity": 2 }
  ]
}
```

---

### Список своих заказов
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/orders?page=1&limit=20` |
| Auth | Bearer Token |

---

### Получить заказ по ID
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/orders/1` |
| Auth | Bearer Token |

> Пользователь видит только свои заказы. Admin видит все.

---

### Отменить заказ (только pending)
| | |
|---|---|
| Method | `PATCH` |
| URL | `{{base_url}}/api/v1/orders/1/cancel` |
| Auth | Bearer Token (владелец заказа) |

---

### Обновить статус заказа (admin / manager)
| | |
|---|---|
| Method | `PATCH` |
| URL | `{{base_url}}/api/v1/orders/1/status` |
| Auth | Bearer Token (admin / manager) |

```json
{
  "status": "confirmed"
}
```

> Допустимые переходы: `pending → confirmed → shipping → delivered`

---

## 8. Пользователи

### Список пользователей (admin)
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/users?page=1&limit=20` |
| Auth | Bearer Token (admin) |

---

### Получить пользователя по ID
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/users/1` |
| Auth | Bearer Token |

---

### Обновить профиль
| | |
|---|---|
| Method | `PUT` |
| URL | `{{base_url}}/api/v1/users/1` |
| Auth | Bearer Token (свой профиль или admin) |

```json
{
  "first_name": "Новое имя",
  "last_name": "Новая фамилия"
}
```

---

### Удалить пользователя (admin)
| | |
|---|---|
| Method | `DELETE` |
| URL | `{{base_url}}/api/v1/users/1` |
| Auth | Bearer Token (admin) |

---

## 9. Seller Dashboard

### Мои продукты (seller / admin)
| | |
|---|---|
| Method | `GET` |
| URL | `{{base_url}}/api/v1/seller/products?page=1&limit=20` |
| Auth | Bearer Token (seller / admin) |

---

## Тестовые аккаунты

| Email | Пароль | Роль |
|---|---|---|
| `admin@example.com` | `Admin1234!` | admin |
| `manager@example.com` | `Manager1234!` | manager |
| `seller@example.com` | `Seller1234!` | seller |
| `customer1@example.com` | `Customer1234!` | customer |
| `customer2@example.com` | `Customer1234!` | customer |
