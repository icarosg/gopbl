# GO-PBL - Sistema de Gerenciamento de Postos de Carregamento de Veículos Elétricos

Este projeto implementa um sistema distribuído para gerenciamento de postos de carregamento de veículos elétricos, utilizando a linguagem Go e comunicação via sockets TCP.

## Arquitetura do Sistema

O sistema é composto por três componentes principais:

1. **Servidor Central**: Gerencia a comunicação entre postos e veículos, mantém registros de postos e veículos, e coordena o processo de reserva de vagas.

2. **Cliente-Posto**: Representa um posto de carregamento que pode cadastrar novos postos, gerenciar filas de espera e atender aos veículos.

3. **Cliente-Veículo**: Representa um veículo elétrico que pode buscar postos disponíveis, solicitar recomendações baseadas em distância e ocupação, e reservar vagas.

## Funcionalidades

### Servidor
- Registro de postos e veículos
- Mediação da comunicação entre postos e veículos
- Cálculo de postos recomendados com base em distância e tempo de espera
- Armazenamento persistente de dados em arquivos JSON

### Cliente-Posto
- Cadastro de novos postos
- Importação de postos existentes
- Gerenciamento da fila de veículos
- Processamento de reservas de vagas

### Cliente-Veículo
- Cadastro de novos veículos
- Importação de veículos existentes
- Busca por postos disponíveis
- Solicitação de recomendação de postos
- Reserva de vagas em postos
- Simulação de deslocamento e carregamento

## Comunicação

A comunicação entre os componentes é realizada através de sockets TCP, utilizando um protocolo próprio baseado em JSON. As mensagens seguem o formato:

```json
{
  "comando": "nome_do_comando",
  "dados": { ... }
}
```

## Requisitos

- Go 1.18 ou superior
- Docker e Docker Compose (para execução em containers)

## Como Executar

### Usando Docker Compose

1. Clone o repositório:
```
git clone https://github.com/icarosg/gopbl.git
cd gopbl
```
2. Contrua as imagens:
```
utilizar o comando docker-compose build para construir as imagens definidas no compose
```
3. Execute o sistema usando Docker Compose:
```
docker-compose up -d
```

4. Interaja com os clientes:
```
# Para interagir com o cliente posto
docker attach gopbl-cliente-posto-1

# Para interagir com o cliente veículo
docker attach gopbl-cliente-veiculo-um-1

# Para interagir com o servidor
docker attach gopbl-main-servidor-1
```

5. Caso queria interagir com mais entidades:
```
#  docker attach gopbl-main-cliente-veiculo-dois-1 para acessar, ver e interagir no veiculo 2
#  docker attach gopbl-main-cliente-veiculo-tres-1 para acessar, ver e interagir no veiculo 3

# docker attach gopbl-main-cliente-posto-um-1 para acessar, ver e interagir no posto 2
```

### Fluxo de Uso

1. No cliente-posto:
   - Escolha a opção 0 para cadastrar um novo posto (informe ID, latitude e longitude)

    - Escolha a opção 1 para listar e importar algum posto

2. No cliente-veículo:
   - Escolha a opção 0 para cadastrar um novo veículo (informe ID, latitude e longitude)

    - Escolha a opção 1 para listar e importar algum veículo

    - Escolha a opção 2 para encontrar o posto recomendado

     - Escolha a opção 3 para reservar uma vaga no posto desejado

   - Escolha a opção 4 para listar os postos disponíveis

## Detalhes de Implementação

- O sistema utiliza um algoritmo de recomendação que considera tanto a distância até o posto quanto o tempo estimado de espera.
- A simulação do deslocamento do veículo ocorre a cada 10 segundos, atualizando a posição do veículo em direção ao posto.
- Os dados de postos e veículos são persistidos em arquivos JSON para permitir a recuperação do estado em caso de reinicialização.
- A comunicação é assíncrona e utiliza goroutines para tratamento paralelo de conexões.

## Autores

Camila Bastos

Guilherme Lopes

Ícaro Gonçalves
