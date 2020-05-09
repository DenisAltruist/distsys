# Домашняя работа по курсу РВ

Реализованы четыре операции для объектов: 

- Создание предмета
- Чтение предмета по коду
- Чтение группы предметов по категории 
- Обновление предмета по коду
- Удаление предмета по коду.

Описание предмета

```
{
    "name": "laptop",
    "code": "42",
    "category": "device"
}
```

Запросы на создание, удаление и редактирование предметов должны быть авторизованными.

![Архитектура](https://app.diagrams.net/?lightbox=1&highlight=0000ff&edit=_blank&layers=1&nav=1&title=Untitled%20Diagram.drawio#R3VrbcqM4EP0aV808hNIFgXn0xNlLzSabXdfsZh4xKLZqsOUV%2BLZfP8KIi4S8JomxM%2BsXo0Y00Of06ZbsAb5d7H4W4Wp%2Bz2OaDBCIdwM8HiAEMSHyK7fsC8sQeIVhJlisJtWGCfuXKiNQ1jWLaapNzDhPMrbSjRFfLmmUabZQCL7Vpz3zRL%2FrKpzRlmEShUnb%2BjeLs7l6CwJq%2By%2BUzeblnSFQZxZhOVkZ0nkY823DhO8G%2BFZwnhVHi90tTfLglXEprvvpyNnqwQRdZl0u%2BMw%2F%2F3X7hYxGD4vt3eTh1%2FENGN%2BgwssmTNbqhQfIS6S%2FT89cupVPne1VKLx%2F1rw8cZMegBrJCdhb7eqT8miWf0%2FmfFV6ko9UOCtOqWhUftF2zjI6WYVRPt5KGslJ82yRyBGUh2G6KoB9ZjsaVw42VGR0dzQWsIqwpCblC5qJvZyiLnCHChTFygqkbY0x9JRt3sS3NIaKV7PKdx16eaCi%2FwIkSE9I3PPljI8%2Flc6mojzxQUZ9kX5sn%2BiKmyT0Kj%2BM9glbxlTIOSegnPK1nBj%2FNq0MYfRtJnLr7%2BtMeqHKfgaIA6JDDIM2xJXKNCFGLukJYs8CsRFTuoxHuWrlYU3CNGWRjEaahSJrmxuBlTER%2Byc5uAEOqAxfpQE4LsGlYZzHBlSjfXP0SAWTr5nDeDAWj0bjljwaAEg9DsWMZqeo3QaqAQSx4FDaBE3CjG30x7Bho%2B7wyNkhXxQPPNfVeIDBUHeR8rWIqLqqqZ%2BmIw85RHcFDTEoItFydSBL9eKv54%2Ffk0T8GU6nLL%2F%2B%2Fo%2BuuS%2BfVlZgejrlz5DJHtQz2YXtTMbYlskmPmfL5KC%2FTN6x7KlxXOSwT9SwzuB8oCXwU3NQXEZIv5lfpM5%2FxAl6HSUCXVMiMMZOgNyg%2BuhpjjBxgLxX%2BTH8d9UPF2LHd0H98XUxQVK6QeMhLiotEPakLaO1BBCBCRUbJqnyflvCqrMrVWZo6ReATWVAXyoDbd15nzLTVWUc0tQZeEJkzqknXVuOq%2BqJ6%2BWprIsIcQNHCstbRYQYLMXBZVsQiPun5CF2Gi2HJ4gpB2YR0zpioPFVOvRPcbb3wtiVyNC9KpON1ssz6da5eTYW3EN8Yd66PdW348vrdUrF%2F3d57bkGM2zl8rLLa2hbX78MY%2BTaMD6OYW35ItF%2BKbJhlPEOsJ4BLWSgVfUsDbR823aX21tvY1tCnSMhH3jGnlkOxo%2FQdJrAYN%2BSRsiWRqgvZErH%2FZZ4pFd48oYKby5632E76l%2BziHvAkGpzW6RrEfdd3zHWxgFEzjAwFL3nWo5sa9UzEbRe4gAQGLzC3pWbRr9rzxhck27DQKeb73uvoxuE2NM9mT%2B%2B9E20tzcUJ2vUo%2BAbFh96h3CRV5zlNF1ZO4wP9zz69vEdF7PA6PF9yz6tvSXsbZ8W2bbMu0iFpWjV2yNfG2detlUCLlibgo5qUbD8WmoBETTlAjgBBm%2FdKIEYyS6j3m8ByLzLZaVk2HtT9aPxs%2BtPA%2FCI6lyGn8jcFiavbJ5cbDrCZ2KgHNb%2FIimm1%2F%2FFwXffAQ%3D%3D)

Коллекция с запросами postman находится в postman/items.postman_collection.json. 

Запуск:
``` docker-compose build && docker-compose up ```
