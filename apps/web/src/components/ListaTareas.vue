<script setup>
import { ref, onMounted } from 'vue'

const tareas = ref([])
const error = ref(null)
const cargando = ref(true)

// Función para obtener los datos desde el backend en Go
const fetchTareas = async () => {
  try {
    const respuesta = await fetch('http://localhost:8080/api/tareas')
    if (!respuesta.ok) throw new Error('Error de red al intentar obtener las tareas')
    tareas.value = await respuesta.json()
  } catch (e) {
    error.value = e.message
  } finally {
    cargando.value = false
  }
}

// Se ejecuta automáticamente cuando el componente se monta en pantalla
onMounted(() => {
  fetchTareas()
})
</script>

<template>
  <div class="contenedor">
    <h2>Gestor de Tareas</h2>
    
    <div v-if="cargando">Cargando tareas...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <ul v-else-if="tareas.length > 0">
      <li v-for="tarea in tareas" :key="tarea.id">
        <span>{{ tarea.titulo }}</span>
        <span class="estado" :class="tarea.estado">{{ tarea.estado }}</span>
      </li>
    </ul>
    
    <div v-else>No hay tareas registradas en la base de datos.</div>
  </div>
</template>

<style scoped>
.contenedor { font-family: Arial, sans-serif; max-width: 400px; margin: auto; }
.error { color: red; }
li { display: flex; justify-content: space-between; padding: 8px; border-bottom: 1px solid #ccc; }
.estado.pendiente { color: orange; font-weight: bold; }
.estado.completado { color: green; font-weight: bold; }
</style>