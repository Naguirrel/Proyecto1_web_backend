# Proyecto1_web_backend
Series Manager – Proyecto #1 Web
Descripción del programa
Series Manager es una aplicación web full stack que permite gestionar un catálogo de series de televisión.
El sistema permite:
•	Crear, editar y eliminar series
•	Visualizar series en formato de tarjetas
•	Buscar series por nombre
•	Ordenar resultados por distintos criterios
•	Paginar resultados dinámicamente
•	Exportar datos a CSV desde el cliente
El frontend está desarrollado con HTML, CSS y JavaScript puro, mientras que el backend expone una API REST para la gestión de datos.

Links de los repositorios y programa
Backend: https://github.com/Naguirrel/Proyecto1_web_backend
Cliente: https://github.com/Naguirrel/Proyecto1_web_cliente
Programa en el server: https://joelsiervas.online/24479/Proyecto1_web_cliente/

Funcionalidades implementadas (Challenge)
Challenges

Criterios Subjetivos
- Calidad visual del cliente: ¿Se ve como una app real o como una tarea? 0 – 30
- Calidad del historial de Git: commits descriptivos, progresión lógica, no un solo commit con todo 0 – 20
- Organización del código: archivos separados con responsabilidades claras, en ambos repositorios 0 – 20

API y Backend
Códigos HTTP correctos en toda la API (201 al crear, 204 al eliminar, 404 si no existe, 400 en input inválido, etc.)	20
Validación server-side con respuestas de error en JSON descriptivas	20
Paginación en GET /series con parámetros ?page= y ?limit=	30
Búsqueda por nombre con ?q=	15
Ordenamiento con ?sort= y ?order=asc|desc	15

Challenges
Exportar la lista de series a CSV — generado manualmente desde JavaScript, sin librerías. El archivo debe descargarse desde el navegador.	20
Exportar la lista de series a Excel (.xlsx) — generado manualmente desde JavaScript, sin librerías de ningún tipo. El archivo debe ser un .xlsx real que abra correctamente en Excel o LibreOffice. Tip: investiguen el formato SpreadsheetML.	30
Sistema de rating — tabla propia en la base de datos, endpoints REST propios (POST /series/:id/rating, GET /series/:id/rating, etc.), y visible en el cliente.	30

Nota total teoríca = 250

Tecnologías utilizadas
•	Frontend: HTML, CSS, JavaScript
•	Backend: API REST
•	Base de datos: (puedes agregar aquí SQLite, PostgreSQL, etc.)
•	Swagger para documentación de la API


Evidencia
<img width="975" height="548" alt="image" src="https://github.com/user-attachments/assets/351b2812-6819-4687-a14a-f83627025eea" />
<img width="975" height="548" alt="image" src="https://github.com/user-attachments/assets/aa51f8a7-0261-48a4-8339-b660feda28ea" />
<img width="975" height="548" alt="image" src="https://github.com/user-attachments/assets/f6fbb5fe-711c-4c56-bda0-ae49d689fddb" />
<img width="975" height="548" alt="image" src="https://github.com/user-attachments/assets/cb939ef2-a36a-4b43-b799-ccec12359dc2" />


Notas adicionales
El proyecto fue desarrollado siguiendo principios de separación de responsabilidades y buenas prácticas en el diseño de APIs REST.
