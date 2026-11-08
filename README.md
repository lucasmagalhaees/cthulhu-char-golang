# cthulhu-char-golang

API de fichas de investigador do Call of Cthulhu (7ª edição), em Go, só com a biblioteca padrão. Reescrita do [cthulhu-caracter-sheet](https://github.com/lucasmagalhaees/cthulhu-caracter-sheet) (Spring Boot): mesmas rotas, mesmo JSON, para o [cthulhu-front](https://github.com/lucasmagalhaees/cthulhu-front) e o Secret Cultist não mudarem nada. O motivo é o tempo de acordar: o container Java levava uns 10 s para subir depois de dormir; o binário Go sobe em milissegundos.

## Rotas

Todas com `Access-Control-Allow-Origin: *`. As listas são guardadas em cache por um dia (`Cache-Control: public, max-age=86400`); fichas e health nunca (`no-store`).

| Rota | O que dá |
|---|---|
| `GET /char-generator/health` | `{"status":"UP","version":"<commit>"}`, sem gerar ficha |
| `GET /char-generator/random-char` | Uma ficha. Opcionais: `stereotype`, `nativeLanguageRegion`, `foreignLanguageRegion` e `seed` (a mesma semente dá a mesma ficha) |
| `GET /char-generator/random-chars?n=10` | De 1 a 30 fichas numa chamada, com as mesmas opções (`n` fora disso dá 400) |
| `GET /char-generator/random-char/stereotype` | Os 20 estereótipos |
| `GET /char-generator/random-char/region/native` | As regiões, com as línguas nativas |
| `GET /char-generator/random-char/region/foreign` | As regiões, com as línguas estrangeiras |
| `GET /` | Redireciona para o health |

## Rodar e testar

```bash
go test -race ./...           # as regras da ficha e o contrato com a resposta gravada do Java
PORT=8080 go run .            # sobe em :8080 (sem PORT, :80)
docker build -t api . && docker run -p 8080:80 api
```

Os testes de `main_test.go` são os mesmos de `CharacterApiTest` do serviço Java: cada regra é conferida em 300 fichas, e `TestContractMatchesTheRecordedJavaAnswer` confere os campos contra uma resposta real gravada em `testdata/`.

## Diferenças do Java, de propósito

- **Taxa de movimento 9:** destreza e força acima do tamanho dão 9, como no livro. No Java esse caso nunca acontecia.
- **Dados:** o d6 vai de 1 a 6. No Java ia de 1 a 5, então a Sorte ficava um pouco mais baixa.
- **Pontos de perícia:** uma perícia do estereótipo não gasta de novo os pontos quando aparece na distribuição geral.

## Deploy

Vercel com runtime de container (`vercel.json`): o `Dockerfile` gera um binário estático numa imagem distroless.
