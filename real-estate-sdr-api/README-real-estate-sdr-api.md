# Real Estate SDR API

Backend REST em Go da solução Real Estate SDR. É responsável por expor
operações de negócio, coordenar casos de uso e acessar os dados
persistidos no PostgreSQL. O serviço Python de IA consulta este backend
por HTTP.

## 1. Arquitetura e responsabilidades

``` text
Python AI / clientes HTTP
          |
          | HTTP + autenticação conforme a rota
          v
Go API — porta 8080
  HTTP handlers / middleware
          |
          v
  Application / casos de uso
          |
          v
  Domain / regras de negócio
          |
          v
  Repositories / Infrastructure / sqlc
          |
          v
  PostgreSQL — porta 5432
```

Responsabilidades principais:

-   Expor endpoints REST para os clientes da solução.
-   Validar requisições e aplicar autenticação conforme configurado.
-   Coordenar casos de uso da aplicação.
-   Aplicar regras de negócio e acessar os dados.
-   Persistir dados no PostgreSQL.
-   Gerenciar a evolução do schema por migrations SQL.
-   Disponibilizar endpoints de saúde/prontidão e logs de diagnóstico,
    conforme a implementação.

## 2. Estrutura do projeto

Estrutura lógica conhecida --- confirme os nomes no checkout atual:

``` text
real-estate-sdr-api/
├── .env                       # configuração local; não versionar
├── .env.example               # exemplo, se presente
├── docker-compose.yml          # infraestrutura/serviços locais
├── scripts/                   # scripts auxiliares, se presentes
└── backend/
    ├── cmd/                   # entrypoints da aplicação
    ├── internal/
    │   ├── application/       # casos de uso/orquestração
    │   ├── domain/            # conceitos e regras do domínio
    │   ├── http/              # handlers, rotas e middleware
    │   ├── infrastructure/    # implementações técnicas e DB
    │   ├── repository/        # acesso a dados
    │   └── observability/     # observabilidade, conforme o código
    ├── migrations/            # migrations SQL versionadas
    ├── queries/               # consultas SQL usadas pelo sqlc
    ├── Dockerfile
    ├── go.mod
    └── sqlc.yaml
```

Para conferir a estrutura:

``` powershell
Get-ChildItem
Get-ChildItem .\backend\cmd -Recurse
Get-ChildItem .\backend\internal -Recurse
```

## 3. Camadas e direção das dependências

-   **Domain:** conceitos e regras centrais do negócio; deve evitar
    dependências de HTTP, banco e provedores externos.
-   **Application:** orquestra os casos de uso e coordena as operações.
-   **HTTP/API:** transforma requisições em chamadas de aplicação e
    respostas HTTP.
-   **Repository/Infrastructure:** implementa persistência e integrações
    técnicas.
-   **Queries/sqlc:** mantém consultas SQL e o código Go tipado
    associado.
-   **Migrations:** registra as alterações do schema em ordem
    versionada.

A direção recomendada é de fora para dentro: handlers e infraestrutura
dependem das abstrações/casos de uso. O domínio não deve depender de
detalhes de transporte ou persistência.

## 4. Pré-requisitos

-   Go instalado e compatível com a versão declarada em
    `backend/go.mod`.
-   Docker Desktop com Docker Compose, caso vá executar o PostgreSQL em
    container.
-   Git.
-   Arquivo `.env` local com as configurações exigidas pelo Compose e
    pela API.

## 5. Configurar o ambiente

Na raiz do repositório:

``` powershell
cd \codes\fiap\8IADT-tech-challenge-fase-5\real-estate-sdr-api
```

Confira o Compose e os arquivos de configuração:

``` powershell
Get-Content .\docker-compose.yml
Get-ChildItem -Force
```

Se existir `.env.example`, copie-o para `.env` e preencha os valores
locais. Não versione credenciais, chaves de API nem senhas.

## 6. Inicializar PostgreSQL e infraestrutura

Na raiz do repositório:

``` powershell
docker compose up -d
docker compose ps
```

A configuração usada durante o desenvolvimento tinha um container
PostgreSQL chamado `real-estate-sdr-postgres`, banco `real_estate_sdr` e
porta local `5432`. Confirme nomes, portas, usuários, senhas, volumes e
health checks no `docker-compose.yml` atual.

Para acompanhar os logs:

``` powershell
docker compose logs -f
```

Para ver apenas as últimas linhas:

``` powershell
docker compose logs --tail 100
```

## 7. Migrations e schema

As migrations SQL ficam em `backend/migrations`. O banco utilizado no
desenvolvimento registrou migrations `001` a `007`; a migration 007
adicionou `properties.status`, com os estados válidos:

-   `AVAILABLE`
-   `RESERVED`
-   `SOLD`
-   `RENTED`
-   `INACTIVE`

A presença de migrations registradas em um banco já existente não
significa que um banco novo será migrado automaticamente. Confirme no
código, scripts ou documentação se a aplicação aplica migrations no
startup ou se há um comando específico.

Antes de alterar schema, remover volumes ou reexecutar migrations
manualmente:

1.  confirme qual banco a API está usando;
2.  confira o histórico de migrations;
3.  faça backup quando houver dados importantes;
4.  evite operações destrutivas como primeira tentativa de diagnóstico.

## 8. Executar a API diretamente

Entre no módulo Go:

``` powershell
cd .\backend
go mod download
```

Execute os testes:

``` powershell
go test -count=1 ./...
```

Descubra o entrypoint disponível:

``` powershell
Get-ChildItem .\cmd -Recurse
```

Execute o pacote real encontrado em `cmd`. Se, por exemplo, houver
`cmd\api`, o comando será:

``` powershell
go run .\cmd\api
```

Se o entrypoint tiver outro nome, ajuste o caminho. A configuração local
usada durante o desenvolvimento publica a API na porta `8080`.

Em outro terminal, teste:

``` powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/ready
```

`/health` verifica a saúde do processo conforme a implementação;
`/ready` foi usado para verificar prontidão, incluindo a conexão com o
banco. Se os caminhos atuais forem diferentes, confira as rotas
registradas no código.

## 9. Executar via Docker Compose

Se o Compose da raiz também construir e iniciar a API, execute da raiz
do repositório:

``` powershell
cd \codes\fiap\8IADT-tech-challenge-fase-5\real-estate-sdr-api
docker compose up --build -d
docker compose ps
docker compose logs -f
```

Use `docker compose down` para parar e remover containers e a rede
criada pelo Compose. Não adicione `-v` sem entender o impacto: essa
opção pode remover volumes e apagar dados persistidos.

## 10. Integração com o serviço Python

A API Go é consumida pelo serviço Python `real-estate-sdr-ai`.

-   Quando Python roda no Windows e Go está em Docker com a porta 8080
    publicada, normalmente o Python pode chamar `http://localhost:8080`.
-   Quando Python e Go estão em containers na mesma rede Compose, o
    Python deve chamar o nome DNS do serviço Go e sua porta interna.
-   Quando Python está em container e Go no host, Docker Desktop para
    Windows normalmente permite `http://host.docker.internal:8080`.
-   Em Azure/produção, configure uma URL apropriada ao ambiente; não use
    `host.docker.internal`.

No ambiente de desenvolvimento é utilizada a variável `API_KEY` para
proteger chamadas da API Go e o cabeçalho `X-API-Key` para enviá-la,
conforme a configuração/middleware. Confirme quais endpoints exigem a
chave.

O backend também participa da correlação de chamadas via
`X-Correlation-ID`, conforme a implementação. Evite registrar valores de
secrets nos logs.

## 11. Diagnóstico rápido

``` powershell
docker compose ps
docker compose logs --tail 100
Test-NetConnection localhost -Port 5432
Test-NetConnection localhost -Port 8080
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/ready
```

Problemas comuns:

-   **PostgreSQL não inicia:** examine os logs do container, variáveis,
    porta ocupada e volume.
-   **API não conecta ao banco:** confira host, porta interna versus
    publicada, credenciais e nome da base.
-   **401/403:** confira `API_KEY`, o cabeçalho `X-API-Key` e as rotas
    protegidas.
-   **Tabela/coluna inexistente:** verifique se as migrations
    necessárias foram aplicadas ao banco correto.
-   **Porta ocupada:** confira os processos e mapeamentos de porta no
    Compose.
-   **Testes falham:** execute `go test -count=1 ./...` dentro de
    `backend` e verifique dependências externas exigidas.

## 12. Segurança e boas práticas

1.  Não versione `.env`, senhas, tokens ou chaves.
2.  Não exponha PostgreSQL diretamente à internet.
3.  Valide entradas e proteja endpoints que não sejam públicos.
4.  Mantenha migrations versionadas e revisadas.
5.  Faça backup antes de operações destrutivas no banco.
6.  Configure timeouts e tratamento de falhas nas integrações HTTP.
7.  Use `X-Correlation-ID` nos logs para rastrear chamadas entre
    serviços.
8.  Em produção, configure TLS, gestão segura de secrets, health checks,
    logs e limites de recursos.
9.  Confirme a compatibilidade do schema antes de publicar uma nova
    versão.

## 13. Checklist de inicialização

-   [ ] `.env` local configurado e fora do Git.
-   [ ] PostgreSQL iniciou e está saudável.
-   [ ] Migrations necessárias foram aplicadas.
-   [ ] API Go iniciou na porta esperada.
-   [ ] `/health` responde.
-   [ ] `/ready` confirma a prontidão.
-   [ ] `go test -count=1 ./...` passa.
-   [ ] O serviço Python consegue acessar a API usando a URL correta.
-   [ ] Uma chamada autenticada foi validada com dados de teste.


## 14. Propriedades de Configuração

- POSTGRES_DB=real_estate_sdr
- POSTGRES_USER=postgres
- POSTGRES_PASSWORD=postgres
- POSTGRES_PORT=5432
- DB_HOST=postgres
- DB_PORT=5432
- DB_NAME=real_estate_sdr
- DB_USER=postgres
- DB_PASSWORD=postgres
- DATABASE_URL=postgres://postgres:postgres@postgres:5432/real_estate_sdr?sslmode=disable
- HTTP_PORT=8080
- OPENAI_API_KEY=...
- OPENAI_MODEL=
- AI_SERVICE_URL=http://host.docker.internal:8001
- API_KEY=
- AI_SERVICE_API_KEY=
- TELEGRAM_BOT_TOKEN=
- TELEGRAM_BOT_USERNAME=real_estate_sdr_ai_bot
