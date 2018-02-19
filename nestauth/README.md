# nestauth

Simple tool convert your:

- Product ID
- Product Secret
- Auth Code

Into an Auth Token.

## Example

1. Create a [Nest Developer Account](developers.nest.com)
2. Create a Product
3. Visit the Authorization URL, authorize and note the auth code

```
./nestauth -client <product id> -secret <product secret> -code <auth code>
token: <some auth token>
expires in: 315360000
```

Use this auth token with the nest_exporter
