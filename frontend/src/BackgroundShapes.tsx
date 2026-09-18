import { useEffect, useRef } from "react";
import Matter from "matter-js";

const SHAPES = ["square", "circle", "triangle", "star"];
const COLORS = [
  "#ff9900",
  "#7cff65",
  "#00a6ff",
  "#ffe45e",
  "#ff4d4d",
  "#a4f58c",
  "#bc85ff",
];

// Seeded random generator
let seed = 42;
function random(min: number, max: number) {
  seed = (seed * 9301 + 49297) % 233280;
  const rnd = seed / 233280;
  return min + rnd * (max - min);
}

interface PhysicsShape {
  id: number;
  type: string;
  color: string;
  size: number;
  body: Matter.Body | null;
  fixedStartX?: number;
}

const STATIC_SHAPES: PhysicsShape[] = (() => {
  seed = 8008; // Reset seed
  const cols = 4;
  const rows = 3;
  const generatedShapes: PhysicsShape[] = [];

  let idCounter = 0;
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      generatedShapes.push({
        id: idCounter++,
        type: SHAPES[Math.floor(random(0, SHAPES.length))],
        color: COLORS[Math.floor(random(0, COLORS.length))],
        size: random(90, 150),
        body: null,
      });
    }
  }
  
  // Explicitly add one extra square to the left
  generatedShapes.push({
    id: idCounter++,
    type: "square",
    color: COLORS[1], // "#7cff65"
    size: 120,
    body: null,
    fixedStartX: 100, // Forces it to spawn on the far left
  });
  
  return generatedShapes;
})();

function ShapeIcon({
  type,
  color,
  size,
}: {
  type: string;
  color: string;
  size: number;
}) {
  if (type === "square") {
    return (
      <div
        className="border-[3px] border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]"
        style={{ width: size, height: size, backgroundColor: color }}
      />
    );
  }

  if (type === "circle") {
    return (
      <div
        className="rounded-full border-[3px] border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]"
        style={{ width: size, height: size, backgroundColor: color }}
      />
    );
  }

  const commonProps = {
    width: size,
    height: size,
    fill: color,
    stroke: "black",
    strokeWidth: 6,
    style: {
      filter: "drop-shadow(4px 4px 0px rgba(0,0,0,1))",
      overflow: "visible",
    },
  };

  if (type === "triangle") {
    return (
      <svg viewBox="0 0 100 100" {...commonProps}>
        <polygon points="50,10 90,90 10,90" strokeLinejoin="round" />
      </svg>
    );
  }

  if (type === "star") {
    return (
      <svg viewBox="0 0 100 100" {...commonProps}>
        <polygon
          points="50,10 61,40 93,40 67,59 77,90 50,71 23,90 33,59 7,40 39,40"
          strokeLinejoin="round"
        />
      </svg>
    );
  }

  return null;
}

export function BackgroundShapes() {
  const containerRef = useRef<HTMLDivElement>(null);
  const elementsRef = useRef<Record<number, HTMLDivElement | null>>({});

  useEffect(() => {
    if (!containerRef.current) return;

    const engine = Matter.Engine.create();
    const world = engine.world;

    const width = containerRef.current.offsetWidth;
    const height = containerRef.current.offsetHeight;

    seed = 9999; // Use a DIFFERENT seed here to prevent correlation with shape types!

    const bodies: Matter.Body[] = [];

    let rightStarAssigned = false;

    // Create bodies
    STATIC_SHAPES.forEach((shape) => {
      const { size, type } = shape;
      // Start them above the screen so they fall in like rain
      let startX = shape.fixedStartX !== undefined ? shape.fixedStartX : random(50, width - 50);
      
      if (type === "star" && !rightStarAssigned) {
        startX = width - random(100, 250); // Force to the right end
        rightStarAssigned = true;
      }
      
      const startY = random(-window.innerHeight, -100);

      let body: Matter.Body;

      // Use rectangle bodies for ALL shapes so the physics bounding box perfectly matches 
      // the DOM wrapper (which is exactly size x size). This prevents SVG centroid offsets 
      // from causing visual clipping on the floor!
      body = Matter.Bodies.rectangle(startX, startY, size, size, {
        restitution: 0.5,
        friction: 0.1,
        density: 0.05,
      });

      // Random initial rotation
      Matter.Body.setAngle(body, random(0, Math.PI * 2));
      
      // Store reference to body
      shape.body = body;
      bodies.push(body);
    });

    Matter.World.add(world, bodies);

    // Create boundaries (Walls, Floor, Ceiling)
    const wallOptions = { isStatic: true, restitution: 0.4, friction: 0.1 };
    
    // Set ground top edge exactly at `height - 1` (the top edge of the 1px blue line)
    // 100/2 = 50, so y - 50 = height - 1 => y = height + 49
    const ground = Matter.Bodies.rectangle(width / 2, height + 49, width * 3, 100, wallOptions);
    const leftWall = Matter.Bodies.rectangle(-50, height / 2, 100, height * 3, wallOptions);
    const rightWall = Matter.Bodies.rectangle(width + 50, height / 2, 100, height * 3, wallOptions);
    const ceiling = Matter.Bodies.rectangle(width / 2, -height * 2, width * 3, 100, wallOptions);

    Matter.World.add(world, [ground, leftWall, rightWall, ceiling]);

    // Add mouse interaction
    const mouse = Matter.Mouse.create(containerRef.current);
    const mouseConstraint = Matter.MouseConstraint.create(engine, {
      mouse: mouse,
      constraint: {
        stiffness: 0.2,
        render: { visible: false },
      },
    });

    // Remove scroll blocking by Matter.js so the user can still scroll freely
    mouse.element.removeEventListener("mousewheel", mouse.mousewheel);
    mouse.element.removeEventListener("DOMMouseScroll", mouse.mousewheel);

    Matter.World.add(world, mouseConstraint);

    const runner = Matter.Runner.create();
    Matter.Runner.run(runner, engine);

    // Sync DOM transforms on every frame without triggering React state updates
    Matter.Events.on(engine, "afterUpdate", () => {
      STATIC_SHAPES.forEach((shape) => {
        if (!shape.body) return;
        
        const domNode = elementsRef.current[shape.id];
        if (domNode) {
          const { x, y } = shape.body.position;
          const angle = shape.body.angle;
          // We calculate the top-left coordinate to translate to, then rotate. 
          // This avoids the rotate() -> translate(-50%) bug which shifts the element off its center!
          domNode.style.transform = `translate(${x - shape.size / 2}px, ${y - shape.size / 2}px) rotate(${angle}rad)`;
        }
      });
    });

    // Handle container resize dynamically to adjust boundaries
    const handleResize = () => {
      if (!containerRef.current) return;
      const newWidth = containerRef.current.offsetWidth;
      const newHeight = containerRef.current.offsetHeight;
      Matter.Body.setPosition(ground, { x: newWidth / 2, y: newHeight + 49 });
      Matter.Body.setPosition(rightWall, { x: newWidth + 50, y: newHeight / 2 });
    };

    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(containerRef.current);

    return () => {
      resizeObserver.disconnect();
      Matter.Runner.stop(runner);
      Matter.Engine.clear(engine);
    };
  }, []);

  return (
    <div
      ref={containerRef}
      className="absolute inset-0 z-0 overflow-hidden pointer-events-auto"
    >
      {STATIC_SHAPES.map((shape) => (
        <div
          key={shape.id}
          ref={(el) => {
            elementsRef.current[shape.id] = el;
          }}
          className="absolute left-0 top-0 will-change-transform cursor-grab active:cursor-grabbing origin-center"
          style={{ width: shape.size, height: shape.size }}
        >
          <ShapeIcon type={shape.type} color={shape.color} size={shape.size} />
        </div>
      ))}
      
      {/* Debug Line exactly at the bottom of the page container */}
      <div className="absolute bottom-0 left-0 w-full h-[1px] bg-transparent z-50 pointer-events-none" />
    </div>
  );
}
