# Real Estate SDR AI

Serviço Python de IA conversacional para atendimento e qualificação de
leads imobiliários. Este serviço recebe mensagens, interpreta intenção e
critérios de busca, gerencia o estado da conversa, integra-se a um LLM e
consulta a API Go para encontrar imóveis compatíveis.

## 1. Arquitetura e responsabilidades

``` text
Lead / canal de entrada (ex.: Telegram)
             |
             v
Python AI — porta 8001
  FastAPI / endpoints
  Agentes e máquina de estados
  Qualificação do lead
  Prompts e integração com LLM
  Cliente HTTP para a API Go
             |
             | HTTP
             v
real-estate-sdr-api — porta 8080
```

Responsabilidades do serviço:

-   Receber e processar mensagens de leads.
-   Identificar intenção e critérios como tipo de imóvel, região,
    orçamento, quartos e urgência.
-   Manter e atualizar o estado conversacional.
-   Orquestrar o fluxo de qualificação e recomendação.
-   Usar o LLM conforme os prompts e as configurações do projeto.
-   Consultar imóveis por meio da API Go.
-   Produzir uma resposta para o canal que iniciou a conversa.

A persistência de negócio deve ficar concentrada na API Go; o serviço
Python não deve acessar diretamente o PostgreSQL de negócio.

## 2. Estrutura do projeto

Estrutura lógica conhecida --- confira os diretórios existentes no
checkout:

``` text
real-estate-sdr-ai/
├── ai/
│   ├── app/
│   │   ├── main.py       # inicialização da aplicação FastAPI
│   │   ├── agent/        # agentes e fluxo conversacional
│   │   ├── api/          # rotas/endpoints HTTP
│   │   ├── clients/      # clientes HTTP e integrações externas
│   │   ├── llm/          # integração com o modelo de linguagem
│   │   ├── models/       # modelos de entrada/saída
│   │   ├── prompt/       # prompts do LLM
│   │   └── services/     # serviços de aplicação e lógica auxiliar
│   └── scripts/
│       └── telegram_bot.py  # bot Telegram, quando presente
├── test/ ou ai/test/      # testes; confirme a localização real
├── requirements.txt       # dependências Python, se presente na raiz
├── .env.example           # exemplo de configuração, se presente
├── .env                   # configuração local; não versionar
└── .venv/                 # ambiente virtual local; não versionar
```

Para conferir a estrutura atual:

``` powershell
Get-ChildItem
Get-ChildItem -Recurse .\ai\app
Get-ChildItem -Recurse .\ai\scripts
```

## 3. Pré-requisitos

-   Python 3.11, conforme o ambiente virtual usado durante o
    desenvolvimento.
-   Acesso à API Go, caso o fluxo precise consultar imóveis.
-   Chave válida do provedor de LLM, caso exigida pela configuração.
-   Git para obter o código.

## 4. Configuração do ambiente

Entre na raiz do repositório:

``` powershell
cd \8IADT-tech-challenge-fase-5\real-estate-sdr-ai
```

Crie o ambiente virtual se ele ainda não existir:

``` powershell
py -3.11 -m venv .venv
```

Ative-o:

``` powershell
.\.venv\Scripts\Activate.ps1
```

Se a ativação for bloqueada pela política do PowerShell, use
`.\.venv\Scripts\python.exe` diretamente nos comandos.

Instale as dependências:

``` powershell
python -m pip install --upgrade pip
python -m pip install -r requirements.txt
```

Se o projeto tiver `.env.example`, crie `.env` a partir dele e preencha
os valores locais. Não publique nem compartilhe o `.env`.

### Variáveis de ambiente

Confira no código quais nomes são lidos. As configurações relevantes
podem incluir:

-   chave do provedor de LLM, por exemplo `OPENAI_API_KEY` se esse for o
    provedor configurado;
-   chave interna do serviço, por exemplo `AI_SERVICE_API_KEY`, se
    usada;
-   URL base da API Go;
-   configurações de logging e demais integrações.

Não coloque valores reais de chaves neste README. A URL da API Go
precisa refletir a topologia local: `http://localhost:8080` se o Go
estiver publicado no host; `http://host.docker.internal:8080` para um
container Python que precise alcançar a API publicada no host; ou o nome
DNS do serviço quando ambos estiverem na mesma rede Docker Compose.

## 5. Executar a API Python

Na raiz do projeto:

``` powershell
.\.venv\Scripts\python.exe -m uvicorn ai.app.main:app --host 0.0.0.0 --port 8001 --reload
```

O `--reload` é indicado para desenvolvimento, não para produção.

Em outro terminal, verifique o endpoint de saúde:

``` powershell
Invoke-RestMethod http://localhost:8001/health
```

Se a aplicação habilitar OpenAPI/Swagger, a documentação normalmente
fica em:

-   `http://localhost:8001/docs`
-   `http://localhost:8001/openapi.json`

Confirme a configuração em `ai/app/main.py`.

## 6. Testes

Execute a suíte a partir da raiz do repositório:

``` powershell
.\.venv\Scripts\python.exe -m pytest -v
```

Ou use a saída resumida:

``` powershell
.\.venv\Scripts\python.exe -m pytest -q
```

Durante o desenvolvimento, a suíte chegou a passar com 68 testes. O
total pode mudar com novas alterações; considere o resultado local atual
como referência.

## 7. Bot Telegram (opcional)

O projeto inclui ou incluiu o script `ai/scripts/telegram_bot.py`.
Confirme que ele existe e confira as variáveis que utiliza antes de
executar:

``` powershell
Get-ChildItem .\ai\scripts
```

Se o caminho estiver correto e o `.env` estiver configurado:

``` powershell
.\.venv\Scripts\python.exe .\ai\scripts\telegram_bot.py
```

O bot pode exigir token do Telegram, URL da API Go e chave de API. Não
imprima tokens em logs nem os versione.

## 8. Integração com a API Go

-   A API Go é o backend responsável pelos dados de imóveis e pelas
    operações de negócio.
-   A URL deve ser configurada no Python conforme o ambiente.
-   Se o Python roda no Windows e o Go está em Docker com a porta 8080
    publicada, `localhost:8080` normalmente funciona.
-   Dentro de um container, `localhost` se refere ao próprio container,
    não a outro serviço.
-   Em Docker Compose, prefira o nome do serviço Go e sua porta interna.
-   Em produção/Azure, não use `host.docker.internal`; configure uma URL
    de rede apropriada e proteja a comunicação.
-   Confira se a rota chamada exige `X-API-Key` ou outro mecanismo de
    autenticação.

## 9. Diagnóstico rápido

``` powershell
Test-NetConnection localhost -Port 8080
Test-NetConnection localhost -Port 8001
Invoke-RestMethod http://localhost:8001/health
```

Problemas comuns:

-   **Falha ao chamar Go:** verifique URL, porta publicada, rede Docker
    e autenticação.
-   **401/403:** confira as chaves e cabeçalhos esperados pelo cliente e
    middleware.
-   **Dependência ausente:** confirme que instalou `requirements.txt` no
    mesmo ambiente usado para iniciar a API.
-   **Erro de LLM:** verifique a chave, o provedor e os limites
    configurados.
-   **Resposta com critérios antigos:** examine o estado conversacional,
    o resultado da extração de critérios e o conteúdo final retornado
    pelo serviço, além dos logs de integração.

## 10. Segurança e operação

1.  Não versione `.env`, tokens ou chaves.
2.  Mantenha `.env.example` sem credenciais reais.
3.  Use timeouts e tratamento de falhas nos clientes HTTP.
4.  Registre logs úteis sem incluir segredos ou dados sensíveis
    desnecessários.
5.  Propague um identificador de correlação entre chamadas quando
    disponível.
6.  Em produção, desative `--reload` e configure health checks, logs e
    gestão segura de secrets.
7.  Não conecte diretamente o serviço Python ao banco de negócio; use a
    API Go.

## 11. Checklist

-   [ ] Python 3.11 disponível.
-   [ ] `.venv` criado e dependências instaladas.
-   [ ] `.env` preenchido localmente e fora do Git.
-   [ ] API Go disponível e acessível a partir do Python.
-   [ ] Endpoint `/health` responde.
-   [ ] Testes Python passam.
-   [ ] Fluxo de ponta a ponta validado com um lead de teste.

## 12. Propriedades de Configuração
- OPENAI_API_KEY=
- OPENAI_MODEL=
- REAL_ESTATE_API_URL=
- DATABASE_URL=postgresql+asyncpg://postgres:postgres@localhost:5432/real_estate_sdr
- FOLLOW_UP_AFTER_MINUTES=1
- FOLLOW_UP_INTERVAL_SECONDS=10
- FOLLOW_UP_MAX_COUNT=1
- AI_SERVICE_API_KEY=
- REAL_ESTATE_API_KEY=
- TELEGRAM_BOT_TOKEN=
- TELEGRAM_BOT_USERNAME=real_estate_sdr_ai_bot
- TELEGRAM_POLLING_TIMEOUT=30
- TELEGRAM_LEAD_MAPPING=
