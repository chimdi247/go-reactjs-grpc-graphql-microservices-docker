// Minimal hand-rolled GraphQL client — no codegen, just fetch() against
// the API gateway's /graphql endpoint. Kept deliberately small and
// dependency-free so there's nothing here that needs a build step.
const GRAPHQL_URL = import.meta.env.VITE_GRAPHQL_URL || 'http://localhost:8080/graphql'

async function gql(query, variables) {
  const token = localStorage.getItem('shopsphere_token')
  const res = await fetch(GRAPHQL_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ query, variables }),
  })
  const json = await res.json()
  if (json.errors && json.errors.length) {
    throw new Error(json.errors.map((e) => e.message).join(', '))
  }
  return json.data
}

export const api = {
  async login(email, password) {
    const data = await gql(
      `mutation Login($email: String!, $password: String!) {
        login(email: $email, password: $password) {
          token
          account { id name email }
        }
      }`,
      { email, password }
    )
    return data.login
  },

  async signup(name, email, password) {
    const data = await gql(
      `mutation CreateAccount($account: AccountInput!) {
        createAccount(account: $account) { id name email }
      }`,
      { account: { name, email, password } }
    )
    return data.createAccount
  },

  async me() {
    const data = await gql(`query { me { id name email } }`)
    return data.me
  },

  async totalAccounts() {
    const data = await gql(`query { totalAccounts }`)
    return data.totalAccounts
  },

  async products(query) {
    const data = await gql(
      `query Products($query: String) {
        products(query: $query, pagination: { skip: 0, take: 50 }) {
          id name description price
        }
      }`,
      { query: query || null }
    )
    return data.products
  },

  async myOrders(accountId) {
    const data = await gql(
      `query Accounts($id: String) {
        accounts(id: $id) {
          id
          orders {
            id
            createdAt
            totalPrice
            products { id name description price quantity }
          }
        }
      }`,
      { id: accountId }
    )
    return data.accounts?.[0]?.orders || []
  },

  async createOrder(accountId, products) {
    const data = await gql(
      `mutation CreateOrder($order: OrderInput!) {
        createOrder(order: $order) {
          id
          createdAt
          totalPrice
        }
      }`,
      {
        order: {
          accountId,
          products: products.map((p) => ({ id: p.id, quantity: p.quantity })),
        },
      }
    )
    return data.createOrder
  },
}
