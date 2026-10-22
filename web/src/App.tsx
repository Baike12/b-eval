import type { ReactNode } from 'react'
import { Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { Agents } from './routes/Agents'
import { Dashboard } from './routes/Dashboard'
import { Metrics } from './routes/Metrics'
import { RunReport } from './routes/RunReport'
import { Runs } from './routes/Runs'
import { TaskDetail } from './routes/TaskSetDetail'
import { TaskSets } from './routes/TaskSets'

// App is the route table.
//
// The routes mirror the platform's objects: a task set holds tasks, a run
// produces a report, and the metric registry is global.
export function App(): ReactNode {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="task-sets" element={<TaskSets />} />
        <Route path="task-sets/:taskSetID/:version/:taskID" element={<TaskDetail />} />
        <Route path="agents" element={<Agents />} />
        <Route path="runs" element={<Runs />} />
        <Route path="runs/:runID" element={<RunReport />} />
        <Route path="metrics" element={<Metrics />} />
      </Route>
    </Routes>
  )
}
