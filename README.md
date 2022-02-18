Go Kit based library to handle authentication using **Microsoft Identity (JWT verification)**, authorization using **Casbin** and API cache using **Redis**.

The library implements middlewares(Decorator patterns) for authentication, authorization and caching.


### HTTP Middlewares 
1. GenericMiddlewareToUpdateEndpointContextForCache - This middleware will update the context with the parameters Cacheable (bool) , Key(string), NotFromCache(bool).
2. JwtMiddlewareForMicrosoftIdentity - This middleware will verify if the JWT token is signed by Microsoft Identity (using public JWK keys verification).

### Endpoint Middlewares 
1.  RedisCacheMiddleware - This middleware will cache the response of the endpoint.
