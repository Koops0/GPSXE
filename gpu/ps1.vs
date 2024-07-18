#version 460 core

in ivec2 vertex_position;
in ivec3 vertex_color;

uniform ivec2 offset

out vec3 color;

void main(){
    ivec2 position = vertex_position + offset

    float xpos = (float(position.x)/512) - 1.0;
    float ypos = 1.0 - (float(position.y)/256);
    gl_Position.xyzw = vec4(xpos, ypos, 0.0, 1.0),

    color = vec3(float(vertex_color.r)/255,
                 float(vertex_color.g)/255,
                 float(vertex_color.b)/255);
}