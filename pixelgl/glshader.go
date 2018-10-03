package pixelgl

import (
	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// GLShader is a type to assist with managing a canvas's underlying
// shader configuration. This allows for customization of shaders on
// a per canvas basis.
type GLShader struct {
	s      *glhf.Shader
	vf, uf glhf.AttrFormat
	vs, fs string

	uniforms []gsUniformAttr

	uniformDefaults struct {
		transform mgl32.Mat3
		colormask mgl32.Vec4
		bounds    mgl32.Vec4
		texbounds mgl32.Vec4
	}
}

type gsUniformAttr struct {
	Name      string
	Type      glhf.AttrType
	value     interface{}
	ispointer bool
}

// reinitialize GLShader data and recompile the underlying gl shader object
func (gs *GLShader) update() {
	gs.uf = nil
	for _, u := range gs.uniforms {
		gs.uf = append(gs.uf, glhf.Attr{
			Name: u.Name,
			Type: u.Type,
		})
	}
	var shader *glhf.Shader
	mainthread.Call(func() {
		var err error
		shader, err = glhf.NewShader(
			gs.vf,
			gs.uf,
			gs.vs,
			gs.fs,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create Canvas, there's a bug in the shader"))
		}
	})

	gs.s = shader
}

// gets the uniform index from GLShader
func (gs *GLShader) getUniform(Name string) int {
	for i, u := range gs.uniforms {
		if u.Name == Name {
			return i
		}
	}
	return -1
}

// AddUniform appends a custom uniform name and value to the shader.
// if the uniform already exists, it will simply be overwritten.
//
// example:
//
//   utime := float32(time.Since(starttime)).Seconds())
//   mycanvas.shader.AddUniform("u_time", &utime)
func (gs *GLShader) AddUniform(Name string, Value interface{}) {
	t, p := getAttrType(Value)
	if loc := gs.getUniform(Name); loc > -1 {
		gs.uniforms[loc].Name = Name
		gs.uniforms[loc].Type = t
		gs.uniforms[loc].ispointer = p
		gs.uniforms[loc].value = Value
		return
	}
	gs.uniforms = append(gs.uniforms, gsUniformAttr{
		Name:      Name,
		Type:      t,
		ispointer: p,
		value:     Value,
	})
}

// Sets up a base shader with everything needed for a Pixel
// canvas to render correctly. The defaults can be overridden
// by simply using the AddUniform function.
func baseShader(c *Canvas) {
	gs := &GLShader{
		vf: defaultCanvasVertexFormat,
		vs: defaultCanvasVertexShader,
		fs: baseCanvasFragmentShader,
	}

	gs.AddUniform("u_transform", &gs.uniformDefaults.transform)
	gs.AddUniform("u_colormask", &gs.uniformDefaults.colormask)
	gs.AddUniform("u_bounds", &gs.uniformDefaults.bounds)
	gs.AddUniform("u_texbounds", &gs.uniformDefaults.texbounds)

	c.shader = gs
}

// Value returns the attribute's concrete value. If the stored value
// is a pointer, we return the dereferenced value.
func (gu *gsUniformAttr) Value() interface{} {
	if !gu.ispointer {
		return gu.value
	}
	switch gu.Type {
	case glhf.Vec2:
		return *gu.value.(*mgl32.Vec2)
	case glhf.Vec3:
		return *gu.value.(*mgl32.Vec3)
	case glhf.Vec4:
		return *gu.value.(*mgl32.Vec4)
	case glhf.Mat2:
		return *gu.value.(*mgl32.Mat2)
	case glhf.Mat23:
		return *gu.value.(*mgl32.Mat2x3)
	case glhf.Mat24:
		return *gu.value.(*mgl32.Mat2x4)
	case glhf.Mat3:
		return *gu.value.(*mgl32.Mat3)
	case glhf.Mat32:
		return *gu.value.(*mgl32.Mat3x2)
	case glhf.Mat34:
		return *gu.value.(*mgl32.Mat3x4)
	case glhf.Mat4:
		return *gu.value.(*mgl32.Mat4)
	case glhf.Mat42:
		return *gu.value.(*mgl32.Mat4x2)
	case glhf.Mat43:
		return *gu.value.(*mgl32.Mat4x3)
	case glhf.Int:
		return *gu.value.(*int32)
	case glhf.Float:
		return *gu.value.(*float32)
	default:
		panic("invalid attrtype")
	}
}

// Returns the type identifier for any (supported) attribute variable type
// and whether or not it is a pointer of that type.
func getAttrType(v interface{}) (glhf.AttrType, bool) {
	switch v.(type) {
	case int32:
		return glhf.Int, false
	case float32:
		return glhf.Float, false
	case mgl32.Vec2:
		return glhf.Vec2, false
	case mgl32.Vec3:
		return glhf.Vec3, false
	case mgl32.Vec4:
		return glhf.Vec4, false
	case mgl32.Mat2:
		return glhf.Mat2, false
	case mgl32.Mat2x3:
		return glhf.Mat23, false
	case mgl32.Mat2x4:
		return glhf.Mat24, false
	case mgl32.Mat3:
		return glhf.Mat3, false
	case mgl32.Mat3x2:
		return glhf.Mat32, false
	case mgl32.Mat3x4:
		return glhf.Mat34, false
	case mgl32.Mat4:
		return glhf.Mat4, false
	case mgl32.Mat4x2:
		return glhf.Mat42, false
	case mgl32.Mat4x3:
		return glhf.Mat43, false
	case *mgl32.Vec2:
		return glhf.Vec2, true
	case *mgl32.Vec3:
		return glhf.Vec3, true
	case *mgl32.Vec4:
		return glhf.Vec4, true
	case *mgl32.Mat2:
		return glhf.Mat2, true
	case *mgl32.Mat2x3:
		return glhf.Mat23, true
	case *mgl32.Mat2x4:
		return glhf.Mat24, true
	case *mgl32.Mat3:
		return glhf.Mat3, true
	case *mgl32.Mat3x2:
		return glhf.Mat32, true
	case *mgl32.Mat3x4:
		return glhf.Mat34, true
	case *mgl32.Mat4:
		return glhf.Mat4, true
	case *mgl32.Mat4x2:
		return glhf.Mat42, true
	case *mgl32.Mat4x3:
		return glhf.Mat43, true
	case *int32:
		return glhf.Int, true
	case *float32:
		return glhf.Float, true
	default:
		panic("invalid AttrType")
	}
}

var defaultCanvasVertexShader = `
#version 330 core

in vec2 position;
in vec4 color;
in vec2 texCoords;
in float intensity;
out vec4 Color;
out vec2 texcoords;
out vec2 glpos;
out float Intensity;

uniform mat3 u_transform;
uniform vec4 u_bounds;

void main() {
	vec2 transPos = (u_transform * vec3(position, 1.0)).xy;
	vec2 normPos = (transPos - u_bounds.xy) / u_bounds.zw * 2 - vec2(1, 1);
	gl_Position = vec4(normPos, 0.0, 1.0);
	Color = color;
	texcoords = texCoords;
	Intensity = intensity;
	glpos = transPos;
}
`

var baseCanvasFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 texcoords;
in float Intensity;

out vec4 fragColor;

uniform vec4 u_colormask;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;

void main() {
	if (Intensity == 0) {
		fragColor = u_colormask * Color;
	} else {
		fragColor = vec4(0, 0, 0, 0);
		fragColor += (1 - Intensity) * Color;
		vec2 t = (texcoords - u_texbounds.xy) / u_texbounds.zw;
		fragColor += Intensity * Color * texture(u_texture, t);
		fragColor *= u_colormask;
	}
}
`
