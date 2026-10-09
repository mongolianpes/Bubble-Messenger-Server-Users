# Bubble Users
## Назначение
Микросервис для управления пользователя в проекте [Bubble](https://github.com/mongolianpes/Bubble-Messenger-Server) <br>
Используется для регистрации, авторизации, получения информации о пользователях, а также используется как TLS

## Функционал
- Отвечает за уровень транспортной безопасности (заменяет TLS)
- Позволяет регистрировать пользователей, проверяя соответствует ли логин, имя пользователя и пароль
- Способствует авторизации пользователей и управляет их сессиями
- Позволяет добавлять аватарки пользователей
- Позволяет получить информацию об авторизованных пользователях по логину и ID устройства
- Позволяет получить информацию о пользователях по ID
- Выполнять поиск среди всех пользователей

## RPC запросы
```
service UsersService {
  rpc TLS(TLSRequest) returns (TLSResponse);
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc AddUserAvatarPath(AddUserAvatarPathRequest) returns (AddUserAvatarPathResponse);
  rpc GetUserAvatarPath(GetUserAvatarPathRequest) returns (GetUserAvatarPathResponse);
  rpc Auth(AuthRequest) returns (AuthResponse);
  rpc GetAuthInfo(GetAuthInfoRequest) returns (GetAuthInfoResponse);
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc GetInfoByID(GetInfoByIDRequest) returns (GetInfoByIDResponse);
}
```
Подробнее в файле users/proto/users.proto

## Переменные окружения
Для корректной работы необходимы переменные окружения:<br>
REDIS_HOST - адрес Redis БД в формате host:port
DB_HOST - адрес БД без порта<br>
DB_PORT - порт БД<br>
DB_USER - пользователь БД<br>
DB_PASSWORD - пароль пользователя БД<br>
DB_NAME - название БД
