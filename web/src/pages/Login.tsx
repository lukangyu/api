import { Button, Card, Form, Input, message } from 'antd'
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
    <div style={{ height: '100vh', display: 'grid', placeItems: 'center' }}>
      <Card title="API 网关管理台登录" style={{ width: 360 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item label="用户名" name="username" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
        </Form>
      </Card>
    </div>
  )
}
