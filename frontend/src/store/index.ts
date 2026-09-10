import { configureStore } from '@reduxjs/toolkit'
import authReducer from './slices/authSlice'
import datasetsReducer from './slices/datasetsSlice'
import searchReducer from './slices/searchSlice'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    datasets: datasetsReducer,
    search: searchReducer,
  },
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
