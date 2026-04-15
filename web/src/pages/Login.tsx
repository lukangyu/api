import { Alert, Button, Card, Col, Form, Input, Row, Space, Typography, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import client from '../api/client'
import { setToken } from '../utils/auth'

export default function LoginPage() {
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const onFinish = async () => {
    try {
      const values = await form.validateFields()
      const resp = await client.post('/auth/login', values)
      setToken(resp.data.token)
      message.success('登录成功')
      navigate('/')
    } catch (e) {
      message.error('登录失败，请检查账号密码')
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        padding: 16,
        background: 'linear-gradient(180deg, #f7fafc 0%, #edf2f7 100%)'
      }}
    >
      <div style={{ width: 'min(980px, 100%)' }}>
        <Row gutter={[24, 24]} align="middle">
          <Col xs={24} md={13}>
            <Card bordered={false} style={{ borderRadius: 16 }}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <div>
                  <Typography.Title level={2} style={{ marginBottom: 8 }}>
                    API 网关管理台
                  </Typography.Title>
                  <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
                    第一次进入时，只需要完成三件事：配置上游、生成员工 Key、复制示例请求做联调。
                  </Typography.Paragraph>
                </div>

                <Alert
                  type="info"
                  showIcon
                  message="首次初始化默认管理员"
                  description="默认账号为 admin，默认密码为 changeme123。这个默认值只适合本地演示；正式环境请在首次初始化前修改 .env 中的 DEFAULT_ADMIN_* 配置。"
                />

                <Card size="small" title="登录后按这个顺序做">
                  <Space direction="vertical" size={12} style={{ width: '100%' }}>
                    <div>
                      <Typography.Text strong>1. 新建上游</Typography.Text>
                      <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
                        先套用服务商模板，再把官方 API Key 填进去。
                      </Typography.Paragraph>
                    </div>
                    <div>
                      <Typography.Text strong>2. 生成员工 Key</Typography.Text>
                      <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
                        给调用方发员工 Key。调用方只接触网关地址和员工 Key，不直接拿上游官方密钥。
                      </Typography.Paragraph>
                    </div>
                    <div>
                      <Typography.Text strong>3. 复制示例请求测试</Typography.Text>
                      <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
                        每个上游页面都可以直接复制专属 curl 示例，用它做第一次联通验证。
                      </Typography.Paragraph>
                    </div>
                  </Space>
                </Card>
              </Space>
            </Card>
          </Col>

          <Col xs={24} md={11}>
            <Card title="登录" style={{ borderRadius: 16 }}>
              <Form form={form} layout="vertical" onFinish={onFinish}>
                <Form.Item label="用户名" name="username" rules={[{ required: true }]}>
                  <Input autoComplete="username" />
                </Form.Item>
                <Form.Item label="密码" name="password" rules={[{ required: true }]}>
                  <Input.Password autoComplete="current-password" />
                </Form.Item>
                <Button type="primary" htmlType="submit" block>
                  登录管理台
                </Button>
              </Form>

              <Typography.Paragraph type="secondary" style={{ marginTop: 16, marginBottom: 0 }}>
                如果你刚启动项目且未改默认配置，可先使用 <code>admin</code> / <code>changeme123</code> 登录验证流程。
              </Typography.Paragraph>
            </Card>
          </Col>
        </Row>
      </div>
    </div>
  )
}
