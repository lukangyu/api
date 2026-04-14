import { Button, Card, Form, Input, InputNumber, Modal, Select, Space, Switch, Table, message } from 'antd'
import { useEffect, useState } from 'react'
import client from '../api/client'

type Upstream = {
  id: number
  name: string
  display_name: string
  base_url: string
  auth_type: string
  auth_key: string
  auth_value: string
  timeout_seconds: number
  strip_prefix: boolean
  extra_headers: string
  is_active: boolean
  description: string
}

function getErrorMessage(err: any): string {
  return err?.response?.data?.error || err?.message || '请求失败'
}

export default function UpstreamsPage() {
  const [items, setItems] = useState<Upstream[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Upstream | null>(null)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const resp = await client.get('/admin/upstreams')
      setItems(resp.data.items || [])
    } catch (err: any) {
      message.error(`加载上游失败：${getErrorMessage(err)}`)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const submit = async () => {
    setSubmitting(true)
    try {
      const values = await form.validateFields()
      if (editing) {
        await client.put(`/admin/upstreams/${editing.id}`, values)
        message.success('更新成功')
      } else {
        await client.post('/admin/upstreams', values)
        message.success('创建成功')
      }
      setOpen(false)
      setEditing(null)
      form.resetFields()
      await load()
    } catch (err: any) {
      message.error(`保存失败：${getErrorMessage(err)}`)
    } finally {
      setSubmitting(false)
    }
  }

  const del = async (id: number) => {
    setSubmitting(true)
    try {
      await client.delete(`/admin/upstreams/${id}`)
      message.success('删除成功')
      await load()
    } catch (err: any) {
      message.error(`删除失败：${getErrorMessage(err)}`)
    } finally {
      setSubmitting(false)
    }
  }

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ auth_type: 'none', strip_prefix: true, timeout_seconds: 120, extra_headers: '{}', is_active: true })
    setOpen(true)
  }

  const applyPreset = (preset: 'producthunt' | 'google' | 'doubao') => {
    if (preset === 'producthunt') {
      form.setFieldsValue({
        name: 'producthunt',
        display_name: 'Product Hunt GraphQL',
        base_url: 'https://api.producthunt.com',
        auth_type: 'bearer',
        auth_key: '',
        auth_value: '',
        strip_prefix: true,
        timeout_seconds: 120,
        extra_headers: '{}',
        description: 'Product Hunt GraphQL API，路径示例：/proxy/producthunt/v2/api/graphql'
      })
      return
    }
    if (preset === 'google') {
      form.setFieldsValue({
        name: 'google',
        display_name: 'Google API',
        base_url: 'https://www.googleapis.com',
        auth_type: 'query',
        auth_key: 'key',
        auth_value: '',
        strip_prefix: true,
        timeout_seconds: 120,
        extra_headers: '{}',
        description: 'Google API，使用 query 参数 key 鉴权'
      })
      return
    }
    form.setFieldsValue({
      name: 'doubao_embedding',
      display_name: 'Doubao Embedding',
      base_url: 'https://ark.cn-beijing.volces.com',
      auth_type: 'none',
      auth_key: '',
      auth_value: '',
      strip_prefix: true,
      timeout_seconds: 120,
      extra_headers: '{}',
      description: '豆包多模态向量，路径示例：/proxy/doubao_embedding/api/v3/embeddings/multimodal，建议 dimensions=2048'
    })
  }

  return (
    <Card
      title="上游 API 管理"
      extra={
        <Button type="primary" onClick={openCreate}>
          新建上游
        </Button>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        dataSource={items}
        columns={[
          { title: '名称', dataIndex: 'name' },
          { title: '显示名', dataIndex: 'display_name' },
          { title: 'Base URL', dataIndex: 'base_url' },
          { title: '认证方式', dataIndex: 'auth_type' },
          {
            title: '启用',
            dataIndex: 'is_active',
            render: (v: boolean) => <Switch checked={v} disabled />
          },
          {
            title: '操作',
            render: (_, row: Upstream) => (
              <Space>
                <Button
                  size="small"
                  onClick={() => {
                    setEditing(row)
                    form.setFieldsValue(row)
                    setOpen(true)
                  }}
                >
                  编辑
                </Button>
                <Button danger size="small" loading={submitting} onClick={() => del(row.id)}>
                  删除
                </Button>
              </Space>
            )
          }
        ]}
      />

      <Modal
        open={open}
        onCancel={() => setOpen(false)}
        onOk={submit}
        confirmLoading={submitting}
        title={editing ? '编辑上游' : '新建上游'}
      >
        {!editing && (
          <Space style={{ marginBottom: 12 }}>
            <Button onClick={() => applyPreset('producthunt')}>套用 Product Hunt</Button>
            <Button onClick={() => applyPreset('google')}>套用 Google</Button>
            <Button onClick={() => applyPreset('doubao')}>套用豆包 Embedding</Button>
          </Space>
        )}

        <Form form={form} layout="vertical">
          <Form.Item label="name" name="name" rules={[{ required: true }]}>
            <Input placeholder="openai / youtube / google / producthunt" />
          </Form.Item>
          <Form.Item label="display_name" name="display_name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="base_url" name="base_url" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="auth_type" name="auth_type" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'none', label: 'none' },
                { value: 'bearer', label: 'bearer' },
                { value: 'header', label: 'header' },
                { value: 'query', label: 'query' }
              ]}
            />
          </Form.Item>
          <Form.Item
            label="auth_key"
            name="auth_key"
            dependencies={['auth_type']}
            rules={[
              ({ getFieldValue }) => ({
                validator(_, value) {
                  const authType = getFieldValue('auth_type')
                  if ((authType === 'header' || authType === 'query') && !value?.trim()) {
                    return Promise.reject(new Error(`auth_type 为 ${authType} 时 auth_key 不能为空`))
                  }
                  return Promise.resolve()
                }
              })
            ]}
          >
            <Input placeholder="query/header 模式必填，例如 key / x-api-key" />
          </Form.Item>
          <Form.Item label="auth_value" name="auth_value">
            <Input.Password placeholder="上游密钥值" />
          </Form.Item>
          <Form.Item label="timeout_seconds" name="timeout_seconds">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="strip_prefix" name="strip_prefix" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="is_active" name="is_active" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="extra_headers(JSON)" name="extra_headers">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item label="description" name="description">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
