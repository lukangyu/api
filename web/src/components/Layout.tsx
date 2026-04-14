import { Layout, Menu, Button } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { clearToken } from '../utils/auth'

const { Header, Sider, Content } = Layout

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider>
        <div style={{ color: '#fff', padding: 16, fontWeight: 700 }}>API 网关</div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          items={[
            { key: '/', label: '仪表盘' },
            { key: '/upstreams', label: '上游配置' },
            { key: '/users', label: '用户管理' },
            { key: '/api-keys', label: 'API Key' },
            { key: '/logs', label: '请求日志' }
          ]}
          onClick={(e) => navigate(e.key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center'
          }}
        >
          <Button
            onClick={() => {
              clearToken()
              navigate('/login')
            }}
          >
            退出登录
          </Button>
        </Header>
        <Content style={{ margin: 16 }}>{children}</Content>
      </Layout>
    </Layout>
  )
}
