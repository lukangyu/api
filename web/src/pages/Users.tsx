import { Button, Card, Form, Input, Modal, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import client from '../api/client'

type User = {
  id: number
  username: string
  display_name: string
  role: string
  is_active: boolean
}

export default function UsersPage() {
  const [items, setItems] = useState<User[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [form] = Form.useForm()

  const load = async () => {
    const resp = await client.get('/admin/users')
    setItems(resp.data.items || [])
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const submit = async () => {
    const values = await form.validateFields()
    if (editing) {
      await client.put(`/admin/users/${editing.id}`, values)
      message.success('更新成功')
    } else {
      await client.post('/admin/users', values)
      message.success('创建成功')
    }
    setOpen(false)
    setEditing(null)
    form.resetFields()
    load()
  }

  const disable = async (id: number) => {
    await client.delete(`/admin/users/${id}`)
    message.success('已停用')
    load()
  }

  return (
    <Card
      title="用户管理"
      extra={
        <Button
          type="primary"
          onClick={() => {
            setEditing(null)
            form.resetFields()
            form.setFieldsValue({ role: 'user' })
            setOpen(true)
          }}
        >
          新建用户
        </Button>
      }
    >
      <Table
        rowKey="id"
        dataSource={items}
        columns={[
          { title: '用户名', dataIndex: 'username' },
          { title: '显示名', dataIndex: 'display_name' },
          { title: '角色', dataIndex: 'role', render: (v: string) => <Tag>{v}</Tag> },
          {
            title: '状态',
            dataIndex: 'is_active',
            render: (v: boolean) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '停用'}</Tag>
          },
          {
            title: '操作',
            render: (_, row: User) => (
              <Space>
                <Button
                  size="small"
                  onClick={() => {
                    setEditing(row)
                    form.setFieldsValue({
                      display_name: row.display_name,
                      role: row.role,
                      is_active: row.is_active
                    })
                    setOpen(true)
                  }}
                >
                  编辑
                </Button>
                <Button danger size="small" onClick={() => disable(row.id)}>
                  停用
                </Button>
              </Space>
            )
          }
        ]}
      />

      <Modal open={open} onCancel={() => setOpen(false)} onOk={submit} title={editing ? '编辑用户' : '新建用户'}>
        <Form form={form} layout="vertical">
          {!editing && (
            <>
              <Form.Item label="用户名" name="username" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item label="密码" name="password" rules={[{ required: true }]}>
                <Input.Password />
              </Form.Item>
            </>
          )}
          <Form.Item label="显示名" name="display_name">
            <Input />
          </Form.Item>
          <Form.Item label="角色" name="role" rules={[{ required: true }]}>
            <Input placeholder="admin / user" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
