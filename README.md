# Guia Técnico de Equivalência CUB • Engenharia & Modelagem Financeira

Landing page técnica e simulador de alta precisão construído em **Go** (`net/http` + templates SSR), **Tailwind CSS v4** (fundo branco com acentos em tons de azul e rosa) e **HTMX**.

Desenvolvido especialmente para **equipes de engenharia de software Node.js (TypeScript / JavaScript)** entenderem e implementarem a regra de negócio da indexação por CUB no mercado imobiliário em fase de obras.

---

## 🎯 O Exemplo Fundamental (R$ 550.000 / CUB 08/2023)

$$\text{Equivalência CUB} = \frac{\text{R\$ } 550.000,00}{\text{R\$ } 2.750,52 \text{ (CUB Agosto/2023)}} = \mathbf{199,9622 \text{ CUBs}}$$

- **CUB-Base Contratual:** R$ 2.750,52 (Agosto/2023).
- **Cotas Imutáveis:** 199,9622 CUBs fixados para todo o ciclo da obra.
- **Valor Atualizado (CUB 10/2026 @ R$ 3.171,96):** R$ 634.272,00 (+15,32%).

---

## 🧩 O Algoritmo em 4 Passos no Node.js

1. **Assinatura & Congelamento:** O CUB do mês da assinatura é gravado no contrato (Prisma/PostgreSQL) e nunca é alterado.
2. **Decomposição em Cotas:** Cada evento financeiro (sinal, parcelas mensais, balões, chaves) é convertido para cotas fixas de CUB via `calcularCotaCUB(valor, cubBase)`.
3. **Série Temporal Sinduscon:** Job Node.js (`node-cron` / BullMQ) ingere o novo índice mensal homologado até o 5º dia útil.
4. **Faturamento no Vencimento:** O endpoint de emissão calcula `atualizarParcela(cotaCUB, cubVigente)`.

---

## 🚀 Como Executar

### 1. Servidor
```bash
./cub-app -port 8080
```
Ou via código-fonte:
```bash
go run ./cmd/server
```
Acesse no navegador: **`http://localhost:8080`**

### 2. Testes Automatizados
```bash
go test -v ./...
```

### 3. Recompilar Tailwind v4 (Paleta em Tons de Azul e Branco)
```bash
npx @tailwindcss/cli -i ./src/input.css -o ./static/css/styles.css --minify
```
